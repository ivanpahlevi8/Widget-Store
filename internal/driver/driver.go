package driver

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// cretate function to init connection
func InitConnection(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)

	// check for an error
	if err != nil {
		log.Println("error when starting connection database")
		return nil, err
	}

	// test for db connection
	err = conn.Ping()

	// check for an error
	if err != nil {
		log.Println("error when trying to ping database")
		return nil, err
	}

	return conn, nil
}
