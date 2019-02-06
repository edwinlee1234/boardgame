package main

import (
	"encoding/json"
	"log"
	"net/http"

	middleware "boardgame_server/middleware"
	pb "boardgame_server/proto"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	redistore "gopkg.in/boj/redistore.v1"
)

var (
	key   = []byte("super-secret-key")
	store *redistore.RediStore
)

// gamecenter address
const (
	gameCenterAddress = "gamecenter:50051"
)

// Redis
var goRedis *redis.Client

// GameCenter
var gameCenter pb.GameCenterClient

func init() {
	connectDb()
	connectRedis()
	connectRedisStore()
	connectGameCenter()
}

func main() {
	r := mux.NewRouter()
	r.Use(middleware.Before)

	// Test
	r.HandleFunc("/", index).Methods("GET", "OPTIONS")
	r.HandleFunc("/test", test).Methods("GET", "OPTIONS")

	// API v1
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/init", initInfo).Methods("GET", "OPTIONS")
	api.HandleFunc("/checkChannel", checkChannel).Methods("GET", "OPTIONS")
	api.HandleFunc("/user/login", loginUser).Methods("POST", "OPTIONS")
	api.HandleFunc("/user/register", registerUser).Methods("POST", "OPTIONS")
	api.HandleFunc("/user/logout", logout).Methods("PUT", "OPTIONS")
	api.HandleFunc("/gamesupport", supportGame).Methods("GET", "OPTIONS")
	api.HandleFunc("/creategame", gameInstance).Methods("PUT", "OPTIONS")
	api.HandleFunc("/roomlist", getRoomList).Methods("GET", "OPTIONS")

	// Game api
	gameAPI := api.PathPrefix("/game").Subrouter()
	gameAPI.HandleFunc("/openplayer", gameOpen).Methods("PUT", "OPTIONS")
	gameAPI.HandleFunc("/roomInfo", gameRoomInfo).Methods("GET", "OPTIONS")
	gameAPI.HandleFunc("/joingame", gameRoomJoin).Methods("PUT", "OPTIONS")
	gameAPI.HandleFunc("/roomClose", gameRoomClose).Methods("PUT", "OPTIONS")
	gameAPI.HandleFunc("/startgame", gameStart).Methods("PUT", "OPTIONS")
	gameAPI.HandleFunc("/info", gameInfo).Methods("GET", "OPTIONS")

	// Jaiput api
	jaipur := gameAPI.PathPrefix("/jaipur").Subrouter()
	jaipur.HandleFunc("/action", jaipurActionProcess).Methods("POST", "OPTIONS")

	err := http.ListenAndServe(":8300", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// 前端進來的init
func initInfo(w http.ResponseWriter, r *http.Request) {
	var res Init
	res.Status = success

	// 登入的判斷 start
	authorization, userID, userName, gameID, err := getSessionUserInfo(r)
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
	res.UserID = userID

	json.NewEncoder(w).Encode(res)
}

func index(w http.ResponseWriter, r *http.Request) {
}

func test(w http.ResponseWriter, r *http.Request) {
}
