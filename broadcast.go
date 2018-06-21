package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	ws "./ws"
)

func pushOpenGame(gameID int) error {
	var openGame OpenGame
	// 用gameID去撈DB
	gameType, state, seat, time := findGameByGameID(gameID)

	rediskey := strconv.Itoa(gameID) + "_players"
	playersList, _ := goRedis.Get(rediskey).Result()
	var playersData Players
	json.Unmarshal([]byte(playersList), &playersData)

	// Redis沒有人，這樣不對
	if len(playersData) <= 0 {
		log.Println("pushOpen ERROR Redis no one player")
		return errors.New("OpenGame Err")
	}

	// 已經開放了
	if state != notOpen {
		log.Println("pushOpen ERROR ", state, " is not 0")
		return errors.New("OpenGame Err")
	}

	// 計算剩下的空位
	seat = seat - len(playersData)

	openGame.Event = "openGame"
	openGame.Data.GameID = gameID
	openGame.Data.GameType = gameType
	openGame.Data.EmptySeat = seat
	openGame.Data.CreateTime = time
	openGame.Data.Players = playersData

	// 轉成json推播
	broadcastData, err := json.Marshal(openGame)
	if err != nil {
		return err
	}

	// 推播到lobby的頻道
	ws.BroadcastChannel(ws.LobbyID, broadcastData)

	// 改變state
	changeGameState(gameID, opening)

	return nil
}

// 有人加入遊戲的推播
func pushChangePlayer(gameID int, players Players) error {
	var changePlayer ChangePlayer
	changePlayer.Event = "ChangePlayer"
	changePlayer.Data = players

	broadcastData, err := json.Marshal(changePlayer)
	if err != nil {
		return err
	}

	ws.BroadcastChannel(gameID, broadcastData)

	return nil
}
