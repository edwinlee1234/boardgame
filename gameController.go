package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// 開新遊戲
// 最後回傳ID
func gameInstance(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	var res Response
	res.Data = map[string][]interface{}{}

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

	// 檢查有沒有正在進行中的遊戲
	session, _ := store.Get(r, "userGame")
	oldGameID, ok := session.Values["gameID"]

	// 有就不給開新的
	if ok {
		log.Println("Exist old gameID: ", oldGameID)
		return
	}

	// DB插新的一局
	id := createGame(game)
	if id == 0 {
		log.Println("Create game ERROR")
		return
	}

	// 寫入Redis
	// 把開這一桌的人推進去Redis
	userUUID := getUserUUID(w, r)
	rediskey := game + strconv.FormatInt(id, 10)
	goRedis.RPush(rediskey, userUUID)

	// 記到玩家session
	session.Values["gameID"] = id
	session.Save(r, w)

	res.Status = success
	res.Data["gameID"] = []interface{}{
		id,
	}

	json.NewEncoder(w).Encode(res)
}

// API 回傳支援的遊戲
func supportGame(w http.ResponseWriter, r *http.Request) {
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

func gameOpen(w http.ResponseWriter, r *http.Request) {

}
