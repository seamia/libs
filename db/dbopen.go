// Copyright 2021 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/elliotchance/sshtunnel"
	_ "github.com/go-sql-driver/mysql"
	libio "github.com/seamia/libs/iox"
)

type Database struct {
	db     *sql.DB
	tunnel *sshtunnel.SSHTunnel
}

func (db *Database) Open(configName string) error {
	cnf, err := libio.LoadJsonAsDictionary(configName)
	if err != nil {
		return err
	}

	if len(cnf["tunnel"]) > 0 {
		// Setup the tunnel, but do not yet start it yet.
		db.tunnel = sshtunnel.NewSSHTunnel(
			// User and host of tunnel server, it will default to port 22 if not specified.
			cnf["tunnel"],

			// Pick ONE of the following authentication methods:
			sshtunnel.PrivateKeyFile(cnf["key"]), // 1. private key
			//		ssh.Password("password"),                            // 2. password
			//		sshtunnel.SSHAgent(),                                // 3. ssh-agent

			// The destination host and port of the actual server.
			cnf["destination"],

			// The local port you want to bind the remote port to.
			// Specifying "0" will lead to a random port.
			"0",
		)

		// You can provide a logger for debugging, or remove this line to // make it silent.
		// tunnel.Log = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)

		// Start the server in the background. You will need to wait a
		// small amount of time for it to bind to the localhost port
		// before you can start sending connections.
		go db.tunnel.Start()
		time.Sleep(100 * time.Millisecond)
	}

	if db.tunnel != nil {
		cnf["host"] = db.tunnel.Local.String()
	}

	const parseTime = "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"

	db.db, err = sql.Open(cnf["driver"], cnf["user"]+":"+cnf["password"]+"@tcp("+cnf["host"]+")/"+cnf["db"]+parseTime)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	return nil
}

func (db Database) DB() *sql.DB {
	return db.db
}

func (db *Database) Close() error {
	if db.tunnel != nil {
		db.tunnel.Close()
		db.tunnel = nil
	}
	if db.db != nil {
		err := db.db.Close()
		db.db = nil
		return err
	}
	return nil
}

/*
func (db Database) Ping() error {
	if db.db != nil {
		return db.db.Ping()
	}
	return errors.New("nil")
}
*/

func (db Database) Ping() error {
	if db.db != nil {

		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()

		err := db.db.PingContext(ctx)
		if err == context.DeadlineExceeded {

		} else if err == context.Canceled {

		}
		return err
	}
	return errors.New("nil")
}
