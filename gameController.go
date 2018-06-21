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
	game := gameParams[0]
	if !checkGameSupport(game) {
		log.Println(game, " is not support")
		return
	}

	// 檢查有沒有正在進行中的遊戲
	session, _ := store.Get(r, "userGame")
	oldGameID, ok := session.Values["gameID"]

	// 有就不給開新的
	if ok {
		res.Status = wrong
		res.Data["oldGameID"] = []interface{}{
			oldGameID,
		}
		json.NewEncoder(w).Encode(res)

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
	rediskey := strconv.Itoa(id) + "_players" // int -> string
	var players Players
	players = append(players, Player{
		ID:   1,
		UUID: userUUID,
		Name: userUUID,
	})
	// jsonencode 加到redis
	playersJSON, _ := json.Marshal(players)
	goRedis.Set(rediskey, playersJSON, 0)

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
	err = pushOpenGame(gameID)

	if err != nil {
		log.Println("PUSH ERROR: ", err)
		return
	}

	var res Response
	res.Status = success
	json.NewEncoder(w).Encode(res)
}

// 取得Room的資料
func gameRoomInfo(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	// 找在場的會員
	var roomInfo RoomInfo
	session, _ := store.Get(r, "userGame")
	gameID, ok := session.Values["gameID"].(int)

	if !ok {
		log.Println("Not found session gameID")
		return
	}

	gameParams := r.URL.Query()["game"]
	if len(gameParams) < 1 {
		log.Println("Url Param 'game' is missing")
		return
	}
	game := gameParams[0]
	if !checkGameSupport(game) {
		log.Println(game, " is not support")
		return
	}

	// 讀Redis
	rediskey := strconv.Itoa(gameID) + "_players"
	playersList, _ := goRedis.Get(rediskey).Result()
	var playersData Players
	json.Unmarshal([]byte(playersList), &playersData)

	// Redis沒有人，這樣不對
	if len(playersData) <= 0 {
		log.Println("pushOpen ERROR Redis no one player")
		return
	}

	// 回傳會員資料
	roomInfo.Data = playersData

	// Room開放了玩家了沒～
	_, state, _, _ := findGameByGameID(gameID)
	if state == notOpen {
		roomInfo.Opening = false
	} else {
		roomInfo.Opening = true
	}

	// 判斷是否場主
	// 順序第一個就是場主
	ownerID := playersData[0].UUID
	userUUID := getUserUUID(w, r)
	if ownerID == userUUID {
		roomInfo.Owner = true
	} else {
		roomInfo.Owner = false
	}

	roomInfo.Status = success

	json.NewEncoder(w).Encode(roomInfo)
}

// API 檢查是否支援這遊戲
func checkGameSupport(game string) bool {
	// 判斷遊戲開關
	if open, exist := gameSupport[game]; !exist || !open {
		return false
	}

	return true
}
