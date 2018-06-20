package ws

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

// Client 連線
// hub是記這個Client是那一個hub
// 同一個hub作推播
type Client struct {
	id   string
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// client傳東西過來
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// 連線設定
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// 一直掛著等Client有東西傳過來
	for {
		// 有東西進來了～
		_, message, err := c.conn.ReadMessage()
		// err處理
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// 格式處理
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// 去推播給同一個hub的人
		c.hub.broadcast <- message
	}
}

// 傳東西去client
func (c *Client) writePump() {
	// 不太懂這個ticker在幹麻(先保留)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Println("Close WS:", c.id)
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		// 要推東西如這個client
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			// 這邊就直接推出去了
			w.Write(message)

			// 不懂這一段的作用，但沒有都可以跑(先保留)
			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 建立ws連線
func serveWs(id string, hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 檢查hub是否已經有連線了
	hub.findClientChan <- id
	oldClient := <-hubFindClientChan
	if oldClient != nil {
		return
	}

	fmt.Println("new Client")
	fmt.Println(hub)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// New新的client
	client := &Client{id: id, hub: hub, conn: conn, send: make(chan []byte, 256)}
	// Hub註冊client地址
	client.hub.register <- client

	// 平行處理Read & Write
	go client.writePump()
	go client.readPump()
}
