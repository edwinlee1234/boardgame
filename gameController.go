package main

import (
	"encoding/json"
	"net/http"
)

func createGame() {

}

// API 回傳支援的遊戲
func supportGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8787")
	// allowOrigin(w, r)
	var res Response
	var resData = map[string][]interface{}{}
	var gameArr []interface{}

	res.Status = success
	for game, open := range gameSupport {
		if open {
			gameArr = append(gameArr, game)
		}
	}
	resData["games"] = gameArr
	res.Data = resData

	json.NewEncoder(w).Encode(res)
}
