package main

var success = "success"
var wrong = "error"

// Response 回應資訊格式
type Response struct {
	Status string                   `json:"status"`
	Data   map[string][]interface{} `json:"data"`
	Error  map[string]interface{}   `json:"error"`
}

// Init 回應資訊格式
type Init struct {
	Status        string `json:"status"`
	Authorization bool   `json:"authorization"`
	UserName      string `json:"userName"`
	GameType      string `json:"gameType"`
	GameID        int    `json:"gameID"`
}

// RoomInfo 回應資訊格式
type RoomInfo struct {
	Status    string  `json:"status"`
	Data      Players `json:"players"`
	Owner     bool    `json:"owner"`
	RoomState string  `json:"roomState"`
	GameID    int     `json:"gameID"`
	GameType  string  `json:"gameType"`
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
	Status     int     `json:"status"`
}

// ChangePlayer 推播遊戲玩家變動的格式
type ChangePlayer struct {
	Event string  `json:"event"`
	Data  Players `json:"data"`
}

// StartGame 推播開始遊戲的格式
type StartGame struct {
	Event string        `json:"event"`
	Data  StartGameData `json:"data"`
}

// StartGameData 推播開始遊戲的格式
type StartGameData struct {
	GameID   int    `json:"gameID"`
	GameType string `json:"gameType"`
}
