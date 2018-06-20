package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	valid "github.com/asaskevich/govalidator"
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
	idInt64 := createGame(game)
	id := int(idInt64) // int64 -> int
	if id == 0 {
		log.Println("Create game ERROR")
		return
	}

	// 寫入Redis
	// 把開這一桌的人推進去Redis
	userUUID := getUserUUID(w, r)
	rediskey := game + strconv.Itoa(id)   // int -> string
	goRedis.RPush(rediskey, "", userUUID) // 第一個要推空白，第二個值才會推進去（好像是）

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
	res.Status = "success"
	res.Data["games"] = gameArr

	json.NewEncoder(w).Encode(res)
}

// 開放玩家進來
func gameOpen(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	IDArrs := r.URL.Query()["id"]
	if len(IDArrs) < 1 {
		log.Println("Url Param 'id' is missing")
		return
	}

	if !valid.IsInt(IDArrs[0]) {
		log.Println("GameID ERROR ID不是數字")
		return
	}

	ParamID, err := strconv.Atoi(IDArrs[0])
	if checkErr("ID 轉換錯誤", err) {
		return
	}

	session, _ := store.Get(r, "userGame")
	gameID, ok := session.Values["gameID"].(int)

	// 比對session的gameID & url帶進來的gameID
	if !ok || gameID != ParamID {
		log.Println("GameID ERROR")
		log.Println(gameID, ParamID)
		return
	}

	// 推播gameID的channel
	pushOpenGame(gameID)
}
