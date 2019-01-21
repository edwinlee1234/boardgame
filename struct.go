package main

var success = "success"
var wrong = "error"

// LobbyChannelID 大廳的channelID
const LobbyChannelID = 1

// Response 回應資訊格式
type Response struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
	Error  map[string]interface{} `json:"error"`
}

func newResponse() *Response {
	return &Response{
		"",
		map[string]interface{}{},
		map[string]interface{}{},
	}
}

// RoomListResponse roomlist API的回傳格式
type RoomListResponse struct {
	Status   string         `json:"status"`
	RoomInfo []OpenGameData `json:"roomlist"`
}

// Init 回應資訊格式
type Init struct {
	Status        string `json:"status"`
	Authorization bool   `json:"authorization"`
	UserID        int32  `json:"userID"`
	UserName      string `json:"userName"`
	GameType      string `json:"gameType"`
	GameID        int32  `json:"gameID"`
}

// RoomInfo 回應資訊格式
type RoomInfo struct {
	Status    string  `json:"status"`
	Data      Players `json:"players"`
	Owner     bool    `json:"owner"`
	RoomState string  `json:"roomState"`
	GameID    int32   `json:"gameID"`
	GameType  string  `json:"gameType"`
}

// Players 回傳會員資訊的格式
type Players []Player

// Player 回傳會員資訊的格式
type Player struct {
	ID   int32 `json:"id"`
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
	GameID     int32   `json:"gameID"`
	Players    Players `json:"players"`
	GameType   string  `json:"gameType"`
	EmptySeat  int32   `json:"emptySeat"`
	CreateTime int32   `json:"time"`
	Status     int32   `json:"status"`
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
	GameID   int32  `json:"gameID"`
	GameType string `json:"gameType"`
}

// KickPlayers 踢掉玩家的推播
type KickPlayers struct {
	Event string  `json:"event"`
	Data  Players `json:"data"`
}

// JaipurActionRequest JaipurActionRequest
type JaipurActionRequest struct {
	GameID   int32        `json:"gameID"`
	GameType string       `json:"gameType"`
	Action   JaipurAction `json:"action"`
}

// JaipurAction JaipurAction
type JaipurAction struct {
	Type             string  `json:"type"`
	Take             int32   `json:"take"`
	Sell             []int32 `json:"sell"`
	SwitchSelfCard   []int32 `json:"switchSelfCard"`
	SwitchTargetCard []int32 `json:"switchTargetCard"`
}
