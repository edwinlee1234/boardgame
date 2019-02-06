package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	model "boardgame_server/model"
)

const wsURL = "http://ws:8000"

// BroadcastRequest 推播的request格式
type BroadcastRequest struct {
	ChannelID int32  `json:"channed_id"`
	Data      []byte `json:"data"`
}

func pushOpenGame(gameID int32) error {
	var openGame OpenGame
	// 用gameID去撈DB
	gameInfo, err := getGameInfoByGameID(gameID)
	if err != nil {
		return err
	}

	gameType, state, seat, time := model.FindGameByGameID(gameID)

	// Redis沒有人，這樣不對
	if len(gameInfo.Players) <= 0 {
		return errors.New("OpenGame Err")
	}

	// 已經開放了
	if state != model.NotOpen {
		return errors.New("OpenGame Err")
	}

	// 改變state db
	if err := model.ChangeGameStateDB(gameID, model.Opening); err != nil {
		return errors.New("change db state Error")
	}
	// 改變state reids
	if err := changeGameInfoRedis(gameID, -1, model.Opening, nil); err != nil {
		return errors.New("change redis state Error")
	}

	// 計算剩下的空位
	seat = seat - int32(len(gameInfo.Players))

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
	broadcastChannel(LobbyChannelID, broadcastData)

	return nil
}

// 推播 post去ws機
func broadcastChannel(channelID int32, data []byte) {
	var req BroadcastRequest
	req.ChannelID = channelID
	req.Data = data

	jsonValue, _ := json.Marshal(req)
	_, err := http.Post(wsURL+"/broadcast", "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		log.Println(err)
	}
}

// 有人加入遊戲的推播
func pushChangePlayer(gameID int32, players Players) error {
	var changePlayer ChangePlayer
	changePlayer.Event = "ChangePlayer"
	changePlayer.Data = players

	broadcastData, err := json.Marshal(changePlayer)
	if err != nil {
		return err
	}

	broadcastChannel(gameID, broadcastData)

	return nil
}

// 開始遊戲
func pushStartGame(gameID int32, gameType string) error {
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

	broadcastChannel(gameID, broadcastData)

	return nil
}

// Room變動的推播
func pushRoomChange(gameID int32) error {
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
	broadcastChannel(LobbyChannelID, broadcastData)

	return nil
}

// 踢掉玩家
func pushKickPlayers(gameID int32, players Players) error {
	var kick KickPlayers

	kick.Event = "Kick"
	kick.Data = players

	broadcastData, err := json.Marshal(kick)
	if err != nil {
		return err
	}

	broadcastChannel(gameID, broadcastData)

	return nil
}

// 場主放棄遊戲
func pushCloseRoom(gameID int32) error {
	var closeRoom CloseRoom
	closeRoom.Event = "CloseRoom"
	closeRoom.GameID = gameID

	broadcastData, err := json.Marshal(closeRoom)
	if err != nil {
		return err
	}

	broadcastChannel(LobbyChannelID, broadcastData)

	return nil
}
