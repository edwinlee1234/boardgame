package main

var success = "success"
var wrong = "error"

// Response 回應資訊格式
type Response struct {
	Status string                   `json:"status"`
	Data   map[string][]interface{} `json:"data"`
}

// RoomInfo 回應資訊格式
type RoomInfo struct {
	Status  string  `json:"status"`
	Data    Players `json:"players"`
	Owner   bool    `json:"owner"`
	Opening bool    `json:"opening"`
}

// Players 回傳會員資訊的格式
type Players []Player

// Player 回傳會員資訊的格式
type Player struct {
	ID   int `json:"id"`
	UUID string
	Name string `json:"name"`
}

// OpenGame 推播開放遊戲的格式
type OpenGame struct {
	Event string       `json:"event"`
	Data  OpenGameData `json:"data"`
}

// OpenGameData 遊戲資料的格式
type OpenGameData struct {
	GameID     int     `json:"gameID"`
	Players    Players `json:"players"`
	GameType   string  `json:"gameType"`
	EmptySeat  int     `json:"emptySeat"`
	CreateTime string  `json:"time"`
}
