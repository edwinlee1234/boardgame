package main

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

func connectDb() {
	err := godotenv.Load()
	checkErr("Error loading .env file", err)

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	db, err = sql.Open(
		"mysql", dbUser+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?charset=utf8mb4")

	checkErr("MySQL Connect", err)
}
