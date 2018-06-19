package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// 開新遊戲
// 最後回傳ID
func gameInstance(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)

	gameParams := r.URL.Query()["game"]
	if len(gameParams) < 1 {
		log.Println("Url Param 'game' is missing")
		return
	}

	// 判斷遊戲開關
	game := gameParams[0]
	if open, exist := gameSupport[game]; !exist || !open {
		log.Println(game, " is not support")
		return
	}

	// DB插新的一局
	id := createGame(game)

	if id == 0 {
		log.Println("Create game ERROR")
		return
	}

	// 寫入Redis

	var res Response
	res.Status = success
	res.Data = map[string][]interface{}{}

	res.Data["gameId"] = []interface{}{
		id,
	}

	json.NewEncoder(w).Encode(res)
}

// API 回傳支援的遊戲
func supportGame(w http.ResponseWriter, r *http.Request) {
	// w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8787")
	allowOrigin(w, r)
	var res Response
	res.Data = map[string][]interface{}{}
	var gameArr []interface{}

	res.Status = success
	for game, open := range gameSupport {
		if open {
			gameArr = append(gameArr, game)
		}
	}
	res.Data["games"] = gameArr

	json.NewEncoder(w).Encode(res)
}
