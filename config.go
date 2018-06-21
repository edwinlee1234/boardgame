package main

// 遊戲狀態
var notOpen = 0 // owner only
var opening = 1 // 開放玩家
var playing = 2 // 遊戲中

var channelSupport = map[string]bool{
	"jaipur": true,
	"lobby":  true,
}

// 支援的遊戲
var gameSupport = map[string]bool{
	"jaipur": true,
}
