package transport

import (
	"net/http"
	"realtimechatserver/internal/server"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// For Dev only 
		return true
	},
}

func ServeWS(h *server.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &server.Client{Hub: h, Conn: conn, Send: make(chan []byte, 16)}

	// Register, then start pumps.
	h.Register <- client
	go client.WritePump()
	go client.ReadPump()
}