package main

import "time"

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
