package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/satori/go.uuid"
)

// 對應client傳進來那一個channel去接連線
func wsInstance(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)

	channelParams := r.URL.Query()["channel"]
	if len(channelParams) < 1 {
		log.Println("Url Param 'channel' is missing")
		return
	}

	// 判斷遊戲開關
	channel := channelParams[0]
	if open, exist := channelSupport[channel]; !exist || !open {
		log.Println(channel, " is not support")
		return
	}
	fmt.Println("channelParams", channelParams)
	userUUID := getUserUUID(w, r)
	fmt.Println("UUID", userUUID)

	var id int
	var hub *Hub

	// 去Group搜尋hub
	if channel == "lobby" {
		id = lobbyID
		group.findHubChan <- id
		hub = <-groupFindHubChan
	}

	fmt.Println("hub", hub)
	// 如果Group沒有這個hub，新增一個
	if hub == nil {
		flag.Parse()
		hub = newHub(id)
		go hub.run()

		group.addHubChan <- hub
	}

	// 新增Client
	serveWs(userUUID, hub, w, r)
}

// 取得UUID
func getUserUUID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "userInfo")
	// 用string的格式取出來
	// *這個用法很重要
	userUUID, ok := session.Values["id"].(string)

	// 如果session沒有，就new一個新的
	if !ok {
		UUID := uuid.Must(uuid.NewV4())
		// UUID轉成string
		userUUID = UUID.String()
		session.Values["id"] = userUUID
		session.Save(r, w)
	}

	return userUUID
}

func createLobby() {
	hub := newHub(lobbyID)
	go hub.run()

	group.addHubChan <- hub
}
