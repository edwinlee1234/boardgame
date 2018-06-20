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

// 	新增一局遊戲
func createGame(game string) int64 {
	stmt, err := db.Prepare("INSERT INTO game_state SET type = ?")
	checkErr("CRUD prepare Error", err)

	val, err := stmt.Exec(game)
	id, _ := val.LastInsertId()
	checkErr("CRUD Exec Error", err)

	return id
}

func findGameByGameID(id int) (gameType string, state int, seat int, time string) {
	row := db.QueryRow("SELECT `type`, `state`, `seat`, `insert_time` game_state FROM `game_state` WHERE `id` = ? LIMIT 1", id)
	err := row.Scan(&gameType, &state, &seat, &time)
	checkErr("find gameType Error:", err)

	return gameType, state, seat, time
}
