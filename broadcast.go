package main

import (
	"encoding/json"
	"log"
	"strconv"

	ws "./ws"
)

func pushOpenGame(GameID int) {
	var openGameData OpenGame
	gameType, state, seat, time := findGameByGameID(GameID)
	rediskey := gameType + strconv.Itoa(GameID)
	players := goRedis.LRange(rediskey, 1, -1)
	playersList := players.Val()

	// 已經開局了
	if state != notOpen {
		log.Println("pushOpen ERROR ", state, " is not 0")
		return
	}

	// Redis沒有人，這樣不對
	if len(playersList) <= 0 {
		log.Println("pushOpen ERROR Redis no one player")
		return
	}

	openGameData.Event = "openGame"
	openGameData.GameID = GameID
	openGameData.GameType = gameType
	openGameData.Seat = seat
	openGameData.CreateTime = time
	openGameData.Players = playersList

	// 轉成json推播
	broadcastData, err := json.Marshal(openGameData)

	if checkErr("pushOpen ERROR Json Marshal ERROR: ", err) {
		return
	}

	// 推播到lobby的頻道
	ws.BroadcastChannel(ws.LobbyID, broadcastData)
}
