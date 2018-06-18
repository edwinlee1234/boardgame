package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

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

// Group
var group *Group

func init() {
	connectDb()
	connectRedis()
	createGroup()
	createLobby()
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/ws", wsInstance).Methods("GET")
	r.HandleFunc("/test", test).Methods("GET")
	r.HandleFunc("/api/gamesupport", supportGame).Methods("GET")

	err := http.ListenAndServe(":8989", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func allowOrigin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8787")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization")
}

func index(w http.ResponseWriter, r *http.Request) {
	supportGame(w, r)
}

func test(w http.ResponseWriter, r *http.Request) {
	for hub, boolan := range group.hubs {
		fmt.Println("hub:")
		fmt.Println(hub)
		fmt.Println(boolan)
		for address, boolan := range hub.clients {
			fmt.Println(address)
			fmt.Println(boolan)
		}
	}
}
