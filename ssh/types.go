// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Dict map[string]string

type Connection struct {
	Info       Dict
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}
