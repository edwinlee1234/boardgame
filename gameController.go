package main

import (
	"encoding/json"
	"errors"
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
	oldGameID, ok := session.Values["gameID"].(int)

	// 有就不給開新的
	if ok {
		rediskey := strconv.Itoa(oldGameID) + "_gameType"
		gameType, _ := goRedis.Get(rediskey).Result()
		if gameType == "" {
			gameType, _, _, _ = findGameByGameID(oldGameID)
		}
		res.Status = wrong
		res.Data["oldGameID"] = []interface{}{
			oldGameID,
		}
		res.Data["gameType"] = []interface{}{
			gameType,
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
	// 記gameType
	rediskey = strconv.Itoa(id) + "_gameType"
	goRedis.Set(rediskey, game, 0)

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
	if r.Method == "OPTIONS" {
		return
	}
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
	if r.Method == "OPTIONS" {
		return
	}
	ParamID, err := checkURLGameID(r)

	if err != nil {
		log.Println(err)
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
	roomInfo.GameID = gameID

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
	gameType, state, _, _ := findGameByGameID(gameID)
	if state == notOpen {
		roomInfo.RoomState = "notOpen"
	} else if state == opening {
		roomInfo.RoomState = "opening"
	} else {
		roomInfo.RoomState = "playing"
	}

	roomInfo.GameType = gameType

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

// 加入Room
func gameRoomJoin(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	var res Response
	res.Data = map[string][]interface{}{}

	// 檢查ID存不存在
	gameID, err := checkURLGameID(r)
	if err != nil {
		log.Println(err)
		return
	}

	// 檢查有沒有正在進行中的遊戲
	session, _ := store.Get(r, "userGame")
	oldGameID, ok := session.Values["gameID"]

	// 有就不給加入新的
	if ok {
		res.Status = wrong
		res.Data["oldGameID"] = []interface{}{
			oldGameID,
		}
		json.NewEncoder(w).Encode(res)

		log.Println("Exist old gameID: ", oldGameID)
		return
	}

	// 寫入Redis
	// 把開這一桌的人推進去Redis
	userUUID := getUserUUID(w, r)
	rediskey := strconv.Itoa(gameID) + "_players" // int -> string
	// 已在的玩家
	playersList, _ := goRedis.Get(rediskey).Result()
	var playersData Players
	json.Unmarshal([]byte(playersList), &playersData)

	// 判斷人數滿了沒
	gameType, state, seat, _ := findGameByGameID(gameID)
	if seat <= len(playersData) {
		res.Status = wrong
		res.Data["msg"] = []interface{}{
			"滿人了",
		}
		json.NewEncoder(w).Encode(res)

		log.Println("滿人了: ", gameID)
		return
	}

	// 判斷是否開放玩家
	if state != opening {
		res.Status = wrong
		res.Data["msg"] = []interface{}{
			"開局了",
		}
		json.NewEncoder(w).Encode(res)

		log.Println("開局了: ", gameID)
		return
	}

	// 插入新的玩家
	playersData = append(playersData, Player{
		ID:   1,
		UUID: userUUID,
		Name: userUUID,
	})

	// jsonencode 加到redis
	playersJSON, _ := json.Marshal(playersData)
	goRedis.Set(rediskey, playersJSON, 0)

	// 記到玩家session
	session.Values["gameID"] = gameID
	session.Save(r, w)

	// 推播
	err = pushChangePlayer(gameID, playersData)

	if err != nil {
		log.Println("推播失敗: ", err)
		return
	}

	res.Status = success
	res.Data["gameID"] = []interface{}{
		gameID,
	}
	res.Data["gameType"] = []interface{}{
		gameType,
	}
	json.NewEncoder(w).Encode(res)

	return
}

// 刪掉Room
func gameRoomClose(w http.ResponseWriter, r *http.Request) {

}

// 開始遊戲
func gameStart(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	gameID, err := checkURLGameID(r)
	if err != nil {
		log.Println("gameID Error: ", err)
		return
	}

	// 不是場主就return
	userUUID := getUserUUID(w, r)
	if !isOwner(gameID, userUUID) {
		return
	}

	err = changeGameState(gameID, playing)
	if err != nil {
		log.Println("DB change Error: ", err)
		return
	}

	// 找gameType
	rediskey := strconv.Itoa(gameID) + "_gameType"
	gameType, _ := goRedis.Get(rediskey).Result()
	if gameType == "" {
		gameType, _, _, _ = findGameByGameID(gameID)
	}

	// 推播開始遊戲
	pushStartGame(gameID, gameType)
	// call gamecenter
	createGameByGameCenter(gameID, gameType)

	var res Response
	res.Status = success
	res.Data = map[string][]interface{}{}
	res.Data["gameID"] = []interface{}{
		gameID,
	}
	res.Data["gameType"] = []interface{}{
		gameType,
	}

	json.NewEncoder(w).Encode(res)
}

// API 檢查是否支援這遊戲
func checkGameSupport(game string) bool {
	// 判斷遊戲開關
	if open, exist := gameSupport[game]; !exist || !open {
		return false
	}

	return true
}

// 檢查URL的GameID參數是否支援這遊戲
func checkURLGameID(r *http.Request) (int, error) {
	IDArrs := r.URL.Query()["id"]
	if len(IDArrs) < 1 {
		return 0, errors.New("Url Param 'id' is missing")
	}

	if !valid.IsInt(IDArrs[0]) {
		return 0, errors.New("GameID ERROR ID不是數字")
	}

	ParamID, err := strconv.Atoi(IDArrs[0])
	if err != nil {
		return 0, errors.New("ID 轉換錯誤")
	}

	return ParamID, nil
}

func isOwner(gameID int, userUUID string) bool {
	// 讀Redis
	rediskey := strconv.Itoa(gameID) + "_players"
	playersList, _ := goRedis.Get(rediskey).Result()
	var playersData Players
	json.Unmarshal([]byte(playersList), &playersData)

	if len(playersData) <= 0 {
		return false
	}

	// 判斷是否場主
	// 順序第一個就是場主
	ownerID := playersData[0].UUID
	if ownerID == userUUID {
		return true
	}

	return false
}
