package main

import (
	"log"
	"os"

	model "./model"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

func connectDb() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if err := model.Connect(dbUser, dbPassword, dbHost, dbPort, dbName); err != nil {
		log.Panic(err)
	}
}
