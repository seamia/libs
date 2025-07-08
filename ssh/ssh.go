// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func (d Dict) String(key, fallback string) string {
	if v, ok := d[key]; ok {
		return v
	}
	return fallback
}

func (c *Connection) value(key string) string {
	return c.Info.String(key, "")
}

func (c *Connection) flag(key string) bool {
	return c.Info.String(key, "") == "yes"
}

func (c *Connection) printf(format string, args ...interface{}) {
	if c.flag("debug") {
		debugPrintf(format, args...)
	}
}

// optional close (if parama call for it)
func (c *Connection) optClose() {
	if c.value("auto.close") == "yes" {
		c.close()
	}
}

func (c *Connection) close() {
	knownConnetionsGuard.Lock()
	defer knownConnetionsGuard.Unlock()

	if c.sftpClient != nil {
		c.sftpClient.Close()
		c.sftpClient = nil
	}
	if c.sshClient != nil {
		c.sshClient.Close()
		c.sshClient = nil
	}
}

func (c *Connection) connect() error {
	if c.sftpClient != nil {
		// already connected
		return nil
	}

	var method ssh.AuthMethod
	var err error

	if keyFileName := c.value("key"); len(keyFileName) != 0 {
		c.printf("using key based auth\n")
		method, err = getSignerFromKey(keyFileName)
		if err != nil {
			onError("error loading key: %v", err)
			return err
		}
	} else if password := c.value("password"); len(password) != 0 {
		c.printf("using password based auth\n")
		method = ssh.Password(password)
	} else {
		return errors.New("no auth method found")
	}

	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// use OpenSSH's known_hosts file if you care about host validation

		c.printf("hostname: %v\n", hostname)
		c.printf("remote:   %v\n", remote)
		c.printf("key:      %v\n", key.Type())

		return nil
	}

	config := &ssh.ClientConfig{
		User: c.value("user"),
		Auth: []ssh.AuthMethod{
			method,
		},
		HostKeyCallback: hostKeyCallback, // ssh.InsecureIgnoreHostKey(),
	}

	/*
		config, err := newSshClient(params.String("user", ""), params.String("key", ""))
		if err != nil {
			onError("error configuring client: %v", err)
			return err
		}
	*/

	addr := c.value("address")
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		onError("error connecting to ssh (%s): %v", addr, err)
		return err
	}
	c.sshClient = client

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		onError("error creating sftp client: %v", err)
		return err
	}
	c.sftpClient = sftp

	return nil
}

func (c *Connection) CreateFile(dstPath string) (*sftp.File, error) {
	if root := c.value("root"); len(root) != 0 {
		dstPath = filepath.Join(root, dstPath)
	}
	dstPath = strings.ReplaceAll(dstPath, "\\", "/")

	// Create the destination file
	return c.sftpClient.Create(dstPath)
}

func getConnection(key string) (*Connection, error) {
	knownConnetionsGuard.Lock()
	defer knownConnetionsGuard.Unlock()

	if knownConnetions == nil {
		var params map[string]Dict

		localConfig := "./ssh.info"
		err := jsonLoadUnmarshal(localConfig, &params)
		if err != nil {

			if homeDir, err := os.UserHomeDir(); err == nil {
				homeBasedConfig := filepath.Join(homeDir, "ssh.info")
				if err := jsonLoadUnmarshal(homeBasedConfig, &params); err != nil {
					onError("error getting ssh info from home folder: %v (%s)", err, homeBasedConfig)
					return nil, err
				}
			} else {
				onError("error getting ssh info: %v (%s)", err, localConfig)
				return nil, err
			}
		}

		knownConnetions = make(map[string]*Connection)
		for key, info := range params {
			conn := &Connection{
				Info: info,
			}
			knownConnetions[key] = conn
		}
	}

	if conn, found := knownConnetions[key]; found {
		if err := conn.connect(); err != nil {
			onError("error connecting to ssh: %v", err)
			return nil, err
		}
		return conn, nil
	}

	return nil, fmt.Errorf("unknown connection name [%s]", key)
}

func getSignerFromKey(keyFileName string) (ssh.AuthMethod, error) {
	// Load private key
	key, err := os.ReadFile(keyFileName)
	if err != nil {
		onError("unable to read private key: %v", err)
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		onError("unable to parse private key: %v", err)
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

func CopyFile(srcPath, dstPath string) error {

	var (
		conn        *Connection
		err         error
		bytesCopied int64
	)

	{
		start := time.Now()
		defer func() {
			took := time.Since(start)
			if conn != nil {
				var rate float64 = float64(bytesCopied) / took.Seconds()
				conn.printf("SSHCopyFile: took %s to complete (rate: %v; bytes: %v)\n", took, int64(rate), bytesCopied)
			}
		}()
	}

	dstParts := strings.Split(dstPath, ":")
	if len(dstParts) != 2 {
		return fmt.Errorf("incorrect destination: %v", dstPath)
	}

	dstKey := dstParts[0]
	conn, err = getConnection(dstKey)
	if err != nil {
		onError("error connecting to ssh (%s): %v", dstKey, err)
		return err
	}

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		onError("error opening source file (%s): %v", srcPath, err)
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := conn.CreateFile(strings.Join(dstParts[1:], ":"))
	if err != nil {
		onError("error creating destination file (%s): %v", dstPath, err)
		return err
	}
	defer dstFile.Close()

	// write to file
	if bytesCopied, err = dstFile.ReadFrom(srcFile); err != nil {
		onError("error copying file: %v", err)
		return err
	}

	conn.optClose()

	return nil
}
