package main

import (
	"encoding/json"
	"errors"
	"net/http"

	ErrorManner "boardgame_server/error"

	"github.com/satori/go.uuid"
)

// 開給ws機用的，去檢查這個channel的合法性
func checkChannel(w http.ResponseWriter, r *http.Request) {
	channelParams := r.URL.Query()["channel"]
	if len(channelParams) < 1 {
		ErrorManner.ErrorRespone(errors.New("channel empty"), DATA_EMPTY, w, 400)
		return
	}

	// 判斷頻道開關
	channel := channelParams[0]
	if open, exist := channelSupport[channel]; !exist || !open {
		ErrorManner.ErrorRespone(errors.New("channel error"), CHANNEL_ERROR, w, 400)
		return
	}

	var res Response
	res.Status = success
	json.NewEncoder(w).Encode(res)
}

// 取得UUID
func getUserUUID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "userInfo")
	// 用string的格式取出來
	// *這個用法很重要
	userUUID, ok := session.Values["uuid"].(string)

	// 如果session沒有，就new一個新的
	if !ok {
		UUID := uuid.Must(uuid.NewV4())
		// UUID轉成string
		userUUID = UUID.String()
		session.Values["uuid"] = userUUID
		session.Save(r, w)
	}

	return userUUID
}
