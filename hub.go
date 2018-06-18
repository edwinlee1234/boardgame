package main

// Hub 用來放連線的地方
type Hub struct {
	id int

	destory chan bool

	clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client

	findClientChan chan string
}

var hubFindClientChan = make(chan *Client)

func newHub(id int) *Hub {
	return &Hub{
		id:             id,
		broadcast:      make(chan []byte),
		destory:        make(chan bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		findClientChan: make(chan string),
	}
}

func (h *Hub) findClient(ID string) {
	for client, _ := range h.clients {
		if client.id == ID {
			hubFindClientChan <- client
			return
		}
	}

	hubFindClientChan <- nil
}

func (h *Hub) run() {
	for {
		select {
		// 這邊會扔地址進來
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				// 不太懂為何要這樣寫
				// 為什麼不是直接client.send <- message
				// 保留例子的寫法
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case clientID := <-h.findClientChan:
			h.findClient(clientID)
		case destory := <-h.destory:
			// 把for迴圈return掉
			if destory {
				return
			}
		}
	}
}
