package main

import "time"

// 遊戲狀態
const notOpen = 0 // owner only
const opening = 1 // 開放玩家
const playing = 2 // 遊戲中
const close = 4   // 關

// 預設空位
const defaultSeat = 2

// redis的時效
const redisGameInfoExpire = time.Hour

var channelSupport = map[string]bool{
	"jaipur": true,
	"lobby":  true,
}

// 支援的遊戲
var gameSupport = map[string]bool{
	"jaipur": true,
}
