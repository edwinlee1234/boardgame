package main

import (
	"database/sql"
	"log"
	"net/http"

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

// MySQL
var db *sql.DB

// Redis
var goRedis *redis.Client

func init() {
	connectDb()
	connectRedis()
	ws.CreateGroup()
	ws.CreateLobby()
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/test", test).Methods("GET")
	r.HandleFunc("/ws", wsInstance).Methods("GET")
	r.HandleFunc("/showChannel", showChannel).Methods("GET")
	r.HandleFunc("/api/gamesupport", supportGame).Methods("GET")
	r.HandleFunc("/api/creategame", gameInstance).Methods("GET")
	r.HandleFunc("/api/game/openplayer", gameOpen).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/game/roomInfo", gameRoomInfo).Methods("GET")
	r.HandleFunc("/api/game/joingame", gameRoomJoin).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/game/roomClose", gameRoomClose).Methods("PUT", "OPTIONS")

	err := http.ListenAndServe(":8989", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func allowOrigin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8787")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
}

func index(w http.ResponseWriter, r *http.Request) {
	getUserUUID(w, r)
}

func test(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
}

func showChannel(w http.ResponseWriter, r *http.Request) {
	ws.CheckAllChannel()
}
