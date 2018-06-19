package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	ws "./ws"

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

	// 判斷頻道開關
	channel := channelParams[0]
	if open, exist := channelSupport[channel]; !exist || !open {
		log.Println(channel, " is not support")
		return
	}
	fmt.Println("channelParams", channelParams)
	userUUID := getUserUUID(w, r)
	fmt.Println("UUID", userUUID)

	var channelID int
	channelIDArrs := r.URL.Query()["id"]
	if len(channelParams) > 1 {
		var err error
		channelIDArr := channelIDArrs[0]
		channelID, err = strconv.Atoi(channelIDArr)

		if checkErr("id Error", err) {
			fmt.Println("create channel error")
			return
		}
	}

	if channel == "lobby" {
		channelID = 1
	}

	// 連線ws
	ws.ConnWs(channelID, userUUID, w, r)
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