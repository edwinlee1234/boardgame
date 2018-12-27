package model

import (
	"database/sql"
)

var db *sql.DB

// Connect 連DB
func Connect(dbUser, dbPassword, dbHost, dbPort, dbName string) error {
	var err error
	db, err = sql.Open(
		"mysql", dbUser+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?charset=utf8mb4")

	if err != nil {
		return err
	}

	return err
}
