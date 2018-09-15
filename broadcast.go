package main

import (
	"encoding/json"
	"errors"

	ws "./ws"
)

func pushOpenGame(gameID int) error {
	var openGame OpenGame
	// 用gameID去撈DB
	gameInfo, err := getGameInfoByGameID(gameID)
	if err != nil {
		return err
	}

	gameType, state, seat, time := findGameByGameID(gameID)

	// Redis沒有人，這樣不對
	if len(gameInfo.Players) <= 0 {
		return errors.New("OpenGame Err")
	}

	// 已經開放了
	if state != notOpen {
		return errors.New("OpenGame Err")
	}

	// 改變state db
	if err := changeGameStateDB(gameID, opening); err != nil {
		return errors.New("change db state Error")
	}
	// 改變state reids
	if err := changeGameInfoRedis(gameID, -1, opening, nil); err != nil {
		return errors.New("change redis state Error")
	}

	// 計算剩下的空位
	seat = seat - len(gameInfo.Players)

	openGame.Event = "openGame"
	openGame.Data.GameID = gameID
	openGame.Data.GameType = gameType
	openGame.Data.EmptySeat = seat
	openGame.Data.CreateTime = time
	openGame.Data.Players = gameInfo.Players

	// 轉成json推播
	broadcastData, err := json.Marshal(openGame)
	if err != nil {
		return err
	}

	// 推播到lobby的頻道
	ws.BroadcastChannel(ws.LobbyID, broadcastData)

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

// 開始遊戲
func pushStartGame(gameID int, gameType string) error {
	var startGame StartGame
	var startGameData StartGameData
	startGameData.GameID = gameID
	startGameData.GameType = gameType

	startGame.Event = "StartGame"
	startGame.Data = startGameData

	broadcastData, err := json.Marshal(startGame)
	if err != nil {
		return err
	}

	ws.BroadcastChannel(gameID, broadcastData)

	return nil
}

// Room變動的推播
func pushRoomChange(gameID int) error {
	// 用gameID去撈DB
	gameInfo, err := getGameInfoByGameID(gameID)
	if err != nil {
		return err
	}

	var openGame OpenGame
	openGame.Event = "RoomChange"
	openGame.Data = gameInfo

	// 轉成json推播
	broadcastData, err := json.Marshal(openGame)
	if err != nil {
		return err
	}

	// 推播到lobby的頻道
	ws.BroadcastChannel(ws.LobbyID, broadcastData)

	return nil
}

// 踢掉玩家
func pushKickPlayers(gameID int, players Players) error {
	var kick KickPlayers

	kick.Event = "Kick"
	kick.Data = players

	broadcastData, err := json.Marshal(kick)
	if err != nil {
		return err
	}

	ws.BroadcastChannel(gameID, broadcastData)

	return nil
}
