package main

var success = "success"

// Response 回應資訊格式
type Response struct {
	Status string                   `json:"status"`
	Data   map[string][]interface{} `json:"data"`
}

// OpenGame 推播開放遊戲的格式
type OpenGame struct {
	Event      string   `json:"event"`
	GameID     int      `json:"gameID"`
	Players    []string `json:"players"`
	GameType   string   `json:"gametype"`
	Seat       int      `json:"seat"`
	CreateTime string   `json:"time"`
}
