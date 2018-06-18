package main

import (
	"os"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

func connectRedis() {
	err := godotenv.Load()
	checkErr("Error loading .env file", err)

	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("REDIS_PORT")

	// 建立連線
	goRedis = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0, // use default DB
	})

	_, err = goRedis.Ping().Result()
	checkErr("Ping Redis Error: ", err)
}
