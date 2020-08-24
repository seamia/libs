// Copyright 2020 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mysql

import (
	"context"
	"database/sql"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

const (
	statementSeparator = ";"
	localFilePrefix    = "file://"
)

func ApplyScript(db *sql.DB, script string) error {
	if db == nil {
		return errors.New("DB is nil")
	}
	if len(script) == 0 {
		return errors.New("script is empty")
	}

	if strings.HasPrefix(script, localFilePrefix) {
		fileName := script[len(localFilePrefix):]
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Printf("Failed to read file (%s) due to (%v)\n", fileName, err)
			return err
		}
		log.Printf("Using script (%s)\n", fileName)
		script = string(data)
	}

	statements := strings.Split(script, statementSeparator)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelfunc()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction (%v)\n", err)
		return err
	}

	for i, statement := range statements {
		line := strings.Trim(statement, " \r\n\t")
		if len(line) > 0 {
			_, err := tx.ExecContext(ctx, line)
			if err != nil {
				log.Printf("Error (%v) on line %i (%s)\n", err, i, statement)
				if err := tx.Rollback(); err != nil {
					log.Printf("Error while rolling back active transaction (%v)\n", err)
				}
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("Error while rolling back active transaction (%v)\n", err)
		return err
	}
	return nil
}
