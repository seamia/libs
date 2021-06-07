package main

import (
	"database/sql"
	"fmt"

	dbssh "github.com/seamia/libs/db"

	_ "github.com/go-sql-driver/mysql"
)

const configName = "db.config"

func main() {
	defer fmt.Println("done.")

	db := dbssh.Database{}
	if err := db.Open(configName); err != nil {
		panic(err)
	}

	showDbVersion(db.DB())

	db.Close()
}

func showDbVersion(db *sql.DB) {
	const stmt = "SHOW VARIABLES where Variable_name like 'Version%';"

	rows, err := db.Query(stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return
		}
		fmt.Printf("%s: %s\n", key, value)
	}
}
