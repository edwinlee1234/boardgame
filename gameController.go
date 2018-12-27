package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	ErrorManner "./error"
	model "./model"

	valid "github.com/asaskevich/govalidator"
)

// 開新遊戲
// 最後回傳ID
func gameInstance(w http.ResponseWriter, r *http.Request) {
	res := newResponse()

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
	timestamp := int32(time.Now().Unix())
	idInt64, err := model.CreateGame(game, defaultSeat, timestamp)
	id := int32(idInt64) // int64 -> int32
	if id == 0 || err != nil {
		ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500)
		return
	}

	// 寫入Redis
	// 把遊戲的資訊全部都寫進去
	userUUID := getUserUUID(w, r)
	rediskey := strconv.Itoa(int(id)) + "_gameInfo" // int32 -> string
	var gameInfo OpenGameData
	// 玩家人數
	var players Players
	players = append(players, Player{
		ID:   userID,
		UUID: userUUID,
		Name: userName,
	})

	gameInfo.CreateTime = timestamp
	// 減掉場主自已
	gameInfo.EmptySeat = defaultSeat - 1
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
	res.Data["gameID"] = id

	json.NewEncoder(w).Encode(res)
}

// API 回傳支援的遊戲
func supportGame(w http.ResponseWriter, r *http.Request) {
	res := newResponse()
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
	res := newResponse()

	// 讀url的參數
	paramGameID, err := checkURLGameID(r)
	if ErrorManner.ErrorRespone(err, DATA_EMPTY, w, 400) {
		return
	}

	// 檢查有沒有正在進行中的遊戲
	authorization, userID, userName, oldGameID, _ := getSessionUserInfo(r)
	// 登入了沒
	if !authorization {
		ErrorManner.ErrorRespone(errors.New("Please login first"), NOT_AUTHORIZATION, w, 400)
		return
	}
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

	// 推播Room
	err = pushChangePlayer(paramGameID, playersData)
	// 不停掉request
	if err != nil {
		ErrorManner.LogsMessage(err, "推播失敗 pushChangePlayer")
	}
	// 推播Lobby
	err = pushRoomChange(paramGameID)
	if err != nil {
		ErrorManner.LogsMessage(err, "推播失敗 pushRoomChange")
	}

	res.Status = success
	res.Data["gameID"] = paramGameID
	res.Data["gameType"] = gameInfo.GameType
	json.NewEncoder(w).Encode(res)

	return
}

// 刪掉Room
func gameRoomClose(w http.ResponseWriter, r *http.Request) {
	gameID, err := checkURLGameID(r)
	if ErrorManner.ErrorRespone(err, DATA_EMPTY, w, 400) {
		return
	}

	// 判斷user是否在這場遊戲中
	_, userID, _, userJoinedGameID, _ := getSessionUserInfo(r)
	if gameID != userJoinedGameID {
		ErrorManner.ErrorRespone(errors.New("User Action error"), USER_ACTION_ERROR, w, 400)
		return
	}

	// 讀遊戲資料
	gameInfo, err := getGameInfoByGameID(gameID)
	if ErrorManner.ErrorRespone(err, UNEXPECT_REDIS_ERROR, w, 500) {
		return
	}

	// 判斷是不是場主，如果是全部人都要被踢出去
	owner, err := isOwner(gameID, userID)
	if ErrorManner.ErrorRespone(err, UNEXPECT_REDIS_ERROR, w, 500) {
		return
	}

	// 不是場主，把自已的session清掉就行
	if !owner {
		// 把這個會員從redis的玩家中刪掉
		var newPlayers Players
		var delUserKey int32
		for index, player := range gameInfo.Players {
			if player.ID == userID {
				delUserKey = int32(index)
			}
		}
		newPlayers = append(gameInfo.Players[:(delUserKey)], gameInfo.Players[(delUserKey+1):]...)
		newSeat := gameInfo.EmptySeat + 1

		err := changeGameInfoRedis(gameID, newSeat, -1, newPlayers)
		if ErrorManner.ErrorRespone(err, UNEXPECT_REDIS_ERROR, w, 500) {
			return
		}

		// lobby change room info的推播
		err = pushRoomChange(gameID)
		// 不停掉request
		if err != nil {
			ErrorManner.LogsMessage(err, "推播失敗 pushRoomChange")
		}

		// 場內人數變動的推播
		err = pushChangePlayer(gameID, newPlayers)
		// 不停掉request
		if err != nil {
			ErrorManner.LogsMessage(err, "推播失敗 pushChangePlayer")
		}
	} else {
		// 場主
		// 改db欄位
		err := model.ChangeGameStateDB(gameID, close)
		if ErrorManner.ErrorRespone(err, UNEXPECT_DB_ERROR, w, 500) {
			return
		}

		// redis delete
		rediskey := strconv.Itoa(int(gameID)) + "_gameInfo"
		goRedis.Del(rediskey)

		// 踢人的推播，玩家全踢
		err = pushKickPlayers(gameID, gameInfo.Players)
		if ErrorManner.ErrorRespone(err, UNEXPECT_BROADCAST_ERROR, w, 500) {
			return
		}

		// TODO
		// 這邊要補一個room delete的event
	}

	// session去掉
	session, _ := store.Get(r, "userInfo")
	session.Values["gameID"] = 0
	session.Save(r, w)

	var res Response
	res.Status = success
	json.NewEncoder(w).Encode(res)
}

// 開始遊戲
func gameStart(w http.ResponseWriter, r *http.Request) {
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
	err = model.ChangeGameStateDB(gameID, playing)
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
	if err := createGameByGameCenter(gameID, gameInfo.GameType, gameInfo.Players); err != nil {
		ErrorManner.ErrorRespone(err, CREATE_GAME_ERROR, w, 500)
		return
	}

	res := newResponse()
	res.Status = success
	res.Data["gameID"] = gameID
	res.Data["gameType"] = gameInfo.GameType

	json.NewEncoder(w).Encode(res)
}

// 取得RoomList
func getRoomList(w http.ResponseWriter, r *http.Request) {
	// 去db讀還開放玩家加入的遊戲
	gameData, err := model.FindAllGameByState(opening)
	if err != nil {
		ErrorManner.ErrorRespone(err, UNEXPECT_DB_ERROR, w, 500)
	}

	// 去讀redis
	// TODO 改成pineline，一次全拿
	var res RoomListResponse
	var roomInfo []OpenGameData
	for _, val := range gameData {
		info, err := getGameInfoByGameID(val.ID)

		if err == nil {
			roomInfo = append(roomInfo, info)
		}
	}

	res.Status = success
	res.RoomInfo = roomInfo

	json.NewEncoder(w).Encode(res)
}

func gameInfo(w http.ResponseWriter, r *http.Request) {
	// 判斷user是否在這場遊戲中
	auth, userID, _, userJoinedGameID, _ := getSessionUserInfo(r)
	if !auth || userJoinedGameID == 0 {
		ErrorManner.ErrorRespone(errors.New("User Action error"), USER_ACTION_ERROR, w, 400)
		return
	}

	gameID, err := checkURLGameID(r)
	if ErrorManner.ErrorRespone(err, DATA_EMPTY, w, 400) {
		return
	}

	if userJoinedGameID != gameID {
		ErrorManner.ErrorRespone(errors.New("User Action error"), USER_ACTION_ERROR, w, 400)
	}

	// 找gameType
	gameInfo, err := getGameInfoByGameID(gameID)
	if ErrorManner.ErrorRespone(err, UNEXPECT_REDIS_ERROR, w, 500) {
		return
	}

	// call gamecenter
	if err := getGameInfo([]int32{userID}, gameID, gameInfo.GameType); err != nil {
		ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500)
		return
	}

	var res Response
	res.Status = success
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
func checkURLGameID(r *http.Request) (int32, error) {
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

	return int32(ParamID), nil
}

func isOwner(gameID int32, userID int32) (bool, error) {
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
