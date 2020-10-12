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

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

const (
	statementSeparator       = ";"
	localFilePrefix          = "file://"
	longestShownScriptLength = 80
)

// (optional) callback func to provide caller-driven error response
type ErrorFilter func(err error, statement string) error

func ApplyScript(db *sql.DB, script string) error {
	return ApplyScriptV2(db, script, nil)
}

func ApplyScriptV2(db *sql.DB, script string, filter ErrorFilter) error {
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
				if filter != nil {
					newErr := filter(err, line)
					if newErr == nil {
						log.WithError(err).Infof("ignoring original error (as instructed by the caller.filter")
						continue
					}
					err = newErr
				}

				if e, ok := err.(*mysql.MySQLError); ok {
					log.WithError(e).Errorf("Error on line %d (%v, %v): (%s)\n", i, e.Number, e.Message, prettyStatement(statement))
				} else {
					log.WithError(err).Errorf("Error on line %d: (%s)\n", i, prettyStatement(statement))
				}

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

func prettyStatement(txt string) string {
	txt = strings.Trim(txt, " \t\r\n")
	if len(txt) > longestShownScriptLength {
		txt = txt[:longestShownScriptLength] + "..."
	}
	return txt
}
