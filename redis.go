package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	redistore "gopkg.in/boj/redistore.v1"
)

func connectRedisStore() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}

	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("REDIS_PORT")

	store, err = redistore.NewRediStore(10, "tcp", host+":"+port, password, key)
	if err != nil {
		panic(err)
	}
}

func connectRedis() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}

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
	if err != nil {
		log.Panic(err)
	}
}

// 用gameID去Redis讀遊戲資料
func getGameInfoByGameID(gameID int32) (OpenGameData, error) {
	rediskey := gameInfoRedisPrefix(gameID)
	infoJSON, err := goRedis.Get(rediskey).Result()
	if err != nil {
		return OpenGameData{}, err
	}

	var info OpenGameData
	err = json.Unmarshal([]byte(infoJSON), &info)
	if err != nil {
		return OpenGameData{}, err
	}

	return info, err
}

// 改Redis gameinfo 值的func
// status && emptySeat 帶-1，程式就會跳過
func changeGameInfoRedis(gameID int32, emptySeat int32, status int32, playersData Players) error {
	gameInfo, err := getGameInfoByGameID(gameID)
	if err != nil {
		return err
	}

	rediskey := gameInfoRedisPrefix(gameID)

	if status != -1 {
		gameInfo.Status = status
	}

	if emptySeat != -1 {
		gameInfo.EmptySeat = emptySeat
	}

	if playersData != nil {
		gameInfo.Players = playersData
	}

	gameInfoJSON, _ := json.Marshal(gameInfo)
	goRedis.Set(rediskey, gameInfoJSON, redisGameInfoExpire)

	return nil
}

func gameInfoRedisPrefix(gameID int32) string {
	return "game_info:" + strconv.Itoa(int(gameID))
}
