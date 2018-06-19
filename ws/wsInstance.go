package ws

import (
	"flag"
	"fmt"
	"net/http"
)

// CheckAllChannel 把現在的hub & client都println出來
func CheckAllChannel() {
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

// CreateLobby 開server的時候就會create一個lobby的hub
func CreateLobby() {
	hub := NewHub(lobbyID)
	go hub.Run()

	group.addHubChan <- hub
}

// ConnWs 連線websocket
func ConnWs(channel string, UUID string, w http.ResponseWriter, r *http.Request) {
	var id int
	var hub *Hub

	// 去Group搜尋hub
	if channel == "lobby" {
		id = lobbyID
		group.findHubChan <- id
		hub = <-groupFindHubChan
	}

	// TODO
	// 用id去找個一局遊戲的hub

	fmt.Println("hub", hub)
	// 如果Group沒有這個hub，新增一個
	if hub == nil {
		flag.Parse()
		hub = NewHub(id)
		go hub.Run()

		group.addHubChan <- hub
	}

	// 新增Client
	serveWs(UUID, hub, w, r)
}
