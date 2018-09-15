package main

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

type dbGameState []dbGameStateData

type dbGameStateData struct {
	ID         int
	Type       string
	Status     int
	Result     string
	Seat       int
	InsertTime string
	UpdateTime string
}

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
func createGame(game string, seat int, insertTime int) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO game_state (type, seat, insert_time) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}

	val, err := stmt.Exec(game, seat, insertTime)
	if err != nil {
		return 0, err
	}

	id, err := val.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// 用GameID去game_state搜尋遊戲資料
func findGameByGameID(id int) (gameType string, state int, seat int, time int) {
	row := db.QueryRow("SELECT `type`, `state`, `seat`, `insert_time` game_state FROM `game_state` WHERE `id` = ? LIMIT 1", id)
	err := row.Scan(&gameType, &state, &seat, &time)
	checkErr("find gameType Error:", err)

	return gameType, state, seat, time
}

// 改變state
func changeGameStateDB(id int, state int) error {
	stmt, _ := db.Prepare("UPDATE `game_state` set `state` = ? where `id` = ?")
	res, _ := stmt.Exec(state, id)

	affect, err := res.RowsAffected()
	if err != nil || affect == 0 {
		return err
	}

	return nil
}

// Get user info by user name
func getUserInfoByUserName(userName string) (id int, name string, password string, err error) {
	row := db.QueryRow("SELECT `id`, `name`, `password` FROM `user` WHERE `name` = ?", userName)
	err = row.Scan(&id, &name, &password)

	return id, name, password, err
}

func regsiterUser(userName, password string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO `user` (`name`,`password`) VALUES (?,?)")

	if err != nil {
		return 0, err
	}

	val, err := stmt.Exec(userName, password)

	if err != nil {
		return 0, err
	}

	id, _ := val.LastInsertId()

	return id, nil
}

// 找出全部有在開的遊戲
func findOpeningGame() (dbGameState, error) {
	var res dbGameState

	rows, err := db.Query(
		"SELECT `id`, `type`, `state`, `seat`, `insert_time`, `update_time` FROM `game_state` WHERE `state` = ?", opening)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		var row dbGameStateData

		err = rows.Scan(&row.ID, &row.Type, &row.Status, &row.Seat, &row.InsertTime, &row.UpdateTime)
		if err != nil {
			return res, err
		}

		res = append(res, row)
	}

	return res, nil
}
