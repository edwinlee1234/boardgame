package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	ErrorManner "./error"

	valid "github.com/asaskevich/govalidator"
)

// 開新遊戲
// 最後回傳ID
func gameInstance(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	if r.Method == "OPTIONS" {
		return
	}
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

	// 會員登入了沒有
	authorization, userID, userName, gameID, err := getSessionUserInfo(r)
	if err != nil || !authorization || userName == "" || userID == 0 {
		ErrorManner.ErrorRespone(errors.New("Please login first"), NOT_AUTHORIZATION, w, 400)
		return
	}

	// 檢查有沒有正在進行中的遊戲
	// 有就不給開新的
	if gameID != 0 {
		ErrorManner.ErrorRespone(errors.New("Exist Playing Game"), EXIST_GAME_NOT_ALLOW_TO_CREATE_NEW_ONE, w, 200)
		return
	}

	// DB插新的一局
	idInt64, err := createGame(game, defaultSeat)
	id := int(idInt64) // int64 -> int
	if id == 0 || err != nil {
		ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500)
		return
	}

	// 寫入Redis
	// 把遊戲的資訊全部都寫進去
	userUUID := getUserUUID(w, r)
	rediskey := strconv.Itoa(id) + "_gameInfo" // int -> string
	var gameInfo OpenGameData
	// 玩家人數
	var players Players
	players = append(players, Player{
		ID:   userID,
		UUID: userUUID,
		Name: userName,
	})

	timestamp := time.Now().Unix()
	gameInfo.CreateTime = strconv.FormatInt(timestamp, 10) // int64 -> string
	gameInfo.EmptySeat = defaultSeat
	gameInfo.GameID = id
	gameInfo.GameType = game
	gameInfo.Players = players
	gameInfo.Status = notOpen

	gameInfoJSON, _ := json.Marshal(gameInfo)
	goRedis.Set(rediskey, gameInfoJSON, redisGameInfoExpire)

	// 記到玩家session
	session, err := store.Get(r, "userInfo")
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

	// 讀url的參數
	ParamID, err := checkURLGameID(r)
	if ErrorManner.ErrorRespone(err, DATA_EMPTY, w, 400) {
		return
	}

	_, _, _, gameID, err := getSessionUserInfo(r)
	if ErrorManner.ErrorRespone(err, USER_ACTION_ERROR, w, 400) {
		return
	}

	// 比對session的gameID & url帶進來的gameID
	if gameID != ParamID {
		ErrorManner.ErrorRespone(errors.New("User Action Error"), USER_ACTION_ERROR, w, 400)
		return
	}

	// 推播gameID的channel
	err = pushOpenGame(gameID)
	if err != nil {
		ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 400)
		return
	}

	var res Response
	res.Status = success
	json.NewEncoder(w).Encode(res)
}

// 取得Room的資料
func gameRoomInfo(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	var roomInfo RoomInfo

	_, userID, _, gameID, err := getSessionUserInfo(r)
	if err != nil || gameID == 0 {
		log.Println(err, gameID)
		ErrorManner.ErrorRespone(errors.New("Session not found"), SESSION_NOT_FOUND, w, 400)
		return
	}

	// 取得遊戲資料
	gameInfo, err := getGameInfoByGameID(gameID)
	if err != nil {
		ErrorManner.ErrorRespone(errors.New("No this game"), GAME_NOT_FOUND, w, 400)
		return
	}

	// 回傳同一局的玩家資料
	roomInfo.Data = gameInfo.Players

	// Room開放了玩家了沒
	if gameInfo.Status == notOpen {
		roomInfo.RoomState = "notOpen"
	} else if gameInfo.Status == opening {
		roomInfo.RoomState = "opening"
	} else {
		roomInfo.RoomState = "playing"
	}

	roomInfo.GameType = gameInfo.GameType
	roomInfo.GameID = gameInfo.GameID

	// 判斷是否場主
	// 順序第一個就是場主
	ownerID := gameInfo.Players[0].ID
	if ownerID == userID {
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

	// 讀url的參數
	paramGameID, err := checkURLGameID(r)
	if ErrorManner.ErrorRespone(err, DATA_EMPTY, w, 400) {
		return
	}

	// 檢查有沒有正在進行中的遊戲
	_, userID, userName, oldGameID, _ := getSessionUserInfo(r)
	// 有就不給加入新的
	if oldGameID != 0 {
		ErrorManner.ErrorRespone(errors.New("Exist Game Not Allow to join new one"), EXIST_GAME_NOT_ALLOW_TO_CREATE_NEW_ONE, w, 200)
		return
	}

	// 寫入Redis
	gameInfo, err := getGameInfoByGameID(paramGameID)
	// 這邊讀不到可能會是user亂帶gameID or redis資料不見了
	if err != nil {
		ErrorManner.ErrorRespone(errors.New("GameID ERROR"), USER_ACTION_ERROR, w, 400)
		return
	}

	// 把開這一桌的人推進去Redis
	userUUID := getUserUUID(w, r)
	playersData := gameInfo.Players
	newEmptySeat := gameInfo.EmptySeat - 1

	// 判斷人數滿了沒
	if newEmptySeat < 0 {
		ErrorManner.ErrorRespone(errors.New("Room has be full"), USER_ACTION_ERROR, w, 400)
		return
	}

	// 判斷是否開放玩家
	if gameInfo.Status != opening {
		ErrorManner.ErrorRespone(errors.New("Game playing"), USER_ACTION_ERROR, w, 400)
		return
	}

	// 插入新的玩家
	playersData = append(playersData, Player{
		ID:   userID,
		UUID: userUUID,
		Name: userName,
	})

	if err := changeGameInfoRedis(paramGameID, newEmptySeat, -1, playersData); err != nil {
		ErrorManner.ErrorRespone(err, USER_ACTION_ERROR, w, 400)
		return
	}

	// 記到玩家session
	session, _ := store.Get(r, "userInfo")
	session.Values["gameID"] = paramGameID
	session.Save(r, w)

	// 推播
	err = pushChangePlayer(paramGameID, playersData)
	// 不停掉request
	if err != nil {
		ErrorManner.LogsMessage(err, "推播失敗 Join Game")
	}

	res.Status = success
	res.Data["gameID"] = []interface{}{
		paramGameID,
	}
	res.Data["gameType"] = []interface{}{
		gameInfo.GameType,
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
	if ErrorManner.ErrorRespone(err, DATA_EMPTY, w, 400) {
		return
	}

	// 不是場主就return
	_, userID, _, _, _ := getSessionUserInfo(r)
	if owner, err := isOwner(gameID, userID); !owner || err != nil {
		return
	}

	// 把遊戲狀態改為開始遊戲
	err = changeGameStateDB(gameID, playing)
	if ErrorManner.ErrorRespone(err, UNEXPECT_DB_ERROR, w, 500) {
		return
	}
	err = changeGameInfoRedis(gameID, -1, playing, nil)
	if ErrorManner.ErrorRespone(err, UNEXPECT_REDIS_ERROR, w, 500) {
		return
	}

	// 找gameType
	gameInfo, err := getGameInfoByGameID(gameID)
	if ErrorManner.ErrorRespone(err, UNEXPECT_REDIS_ERROR, w, 500) {
		return
	}

	// 推播開始遊戲
	pushStartGame(gameID, gameInfo.GameType)
	// call gamecenter
	// createGameByGameCenter(gameID, gameInfo.GameType)

	var res Response
	res.Status = success
	res.Data = map[string][]interface{}{}
	res.Data["gameID"] = []interface{}{
		gameID,
	}
	res.Data["gameType"] = []interface{}{
		gameInfo.GameType,
	}

	json.NewEncoder(w).Encode(res)
}

// 取得RoomList
func getRoomList(w http.ResponseWriter, r *http.Request) {

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

func isOwner(gameID int, userID int) (bool, error) {
	gameInfo, err := getGameInfoByGameID(gameID)

	if err != nil {
		return false, err
	}

	// 判斷是否場主
	// 順序第一個就是場主
	ownerID := gameInfo.Players[0].ID
	if userID == ownerID {
		return true, nil
	}

	return false, nil
}
