package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	pb "./proto"
	ws "./ws"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

// gamecenter address
const (
	gameCenterAddress = "gamecenter:50051"
)

// MySQL
var db *sql.DB

// Redis
var goRedis *redis.Client

// GameCenter
var gameCenter pb.GameCenterClient

func init() {
	connectDb()
	connectRedis()
	connectGameCenter()
	ws.CreateGroup()
	ws.CreateLobby()
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/test", test).Methods("GET")
	r.HandleFunc("/ws", wsInstance).Methods("GET")
	r.HandleFunc("/showChannel", showChannel).Methods("GET")
	r.HandleFunc("/init", initInfo).Methods("get", "GET")
	r.HandleFunc("/api/user/login", loginUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/user/register", registerUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/gamesupport", supportGame).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/creategame", gameInstance).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/game/openplayer", gameOpen).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/game/roomInfo", gameRoomInfo).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/game/joingame", gameRoomJoin).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/game/roomClose", gameRoomClose).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/game/startgame", gameStart).Methods("PUT", "OPTIONS")

	err := http.ListenAndServe(":8300", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func allowOrigin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8787")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header, x-xsrf-token")
}

// 前端進來的init
func initInfo(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	var res Init
	res.Status = success

	// 登入的判斷
	session, _ := store.Get(r, "userInfo")
	authorization, ok := session.Values["login"].(bool)
	if ok {
		res.Authorization = authorization
	} else {
		res.Authorization = false
	}

	userName, ok := session.Values["userName"].(string)
	if ok && userName != "" {
		res.UserName = userName
	} else {
		res.UserName = "Guest"
	}

	if !res.Authorization {
		json.NewEncoder(w).Encode(res)
	}

	// 檢查有沒有已加入的遊戲
	session, _ = store.Get(r, "userGame")
	gameID, ok := session.Values["gameID"].(int)

	if gameID != 0 && ok {
		rediskey := strconv.Itoa(gameID) + "_gameType"
		gameType, err := goRedis.Get(rediskey).Result()

		if err == nil && gameType != "" {
			res.GameType = gameType
			res.GameID = gameID
		}
	}

	json.NewEncoder(w).Encode(res)
}

func index(w http.ResponseWriter, r *http.Request) {
	getUserUUID(w, r)
}

func test(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	rediskey := "afasf_gameType"
	gameType, err := goRedis.Get(rediskey).Result()
	if gameType == "" {
		fmt.Println("!!")
	}
	fmt.Println(err)
	fmt.Println(gameType)

	session, _ := store.Get(r, "balbal")
	data, ok := session.Values["gameID"].(int)

	fmt.Println(data, ok)
}

func showChannel(w http.ResponseWriter, r *http.Request) {
	ws.CheckAllChannel()
}
