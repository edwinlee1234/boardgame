package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

	r.HandleFunc("/", index).Methods("GET", "OPTIONS")
	r.HandleFunc("/test", test).Methods("GET", "OPTIONS")
	r.HandleFunc("/ws", wsInstance).Methods("GET", "OPTIONS")
	r.HandleFunc("/showChannel", showChannel).Methods("GET", "OPTIONS")
	r.HandleFunc("/init", initInfo).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/user/login", loginUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/user/register", registerUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/gamesupport", supportGame).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/creategame", gameInstance).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/roomlist", getRoomList).Methods("GET", "OPTIONS")
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
	if r.Method == "OPTIONS" {
		return
	}
	var res Init
	res.Status = success

	// 登入的判斷 start
	authorization, _, userName, gameID, err := getSessionUserInfo(r)
	if err != nil && !authorization {
		res.Authorization = false
		json.NewEncoder(w).Encode(res)

		return
	}
	res.Authorization = authorization
	res.UserName = userName

	// 檢查有沒有已加入的遊戲
	gameInfo, err := getGameInfoByGameID(gameID)
	if err != nil {
		// TODO 這個是防止，redis已經死了，但user的gameID session還在，一律都變0好了，這邊可以會造成太多次的修改
		session, _ := store.Get(r, "userInfo")
		session.Values["gameID"] = 0
		session.Save(r, w)

		json.NewEncoder(w).Encode(res)
		return
	}

	res.GameType = gameInfo.GameType
	res.GameID = gameInfo.GameID

	json.NewEncoder(w).Encode(res)
}

func index(w http.ResponseWriter, r *http.Request) {
}

func test(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	authorization, userID, userName, gameID, err := getSessionUserInfo(r)
	gameInfo, err := getGameInfoByGameID(gameID)
	fmt.Println(authorization, userID, userName, gameID, err)
	fmt.Println(gameInfo, err)
}

func showChannel(w http.ResponseWriter, r *http.Request) {
	ws.CheckAllChannel()
}
