package server

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 1024
)

type Client struct {
	Hub *Hub
	Conn *websocket.Conn
	Send chan []byte
	Partner *Client
	LastActive time.Time
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	
	c.LastActive = time.Now()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			// connection closed or read error
			break
		}

		msgStr := string(message) 
		if msgStr == "/next" {
			// Skip request → unregister, then re-register.
			c.Hub.Unregister <- c
			c.Hub.Register <- c
			continue
		}

		// Relay only to partner if paired
		if p := c.Partner; p != nil {
			select {
			case p.Send <- message:
			default:
			}
		}
	}
}


func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select  {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}

		if time.Since(c.LastActive) > 5*time.Minute {
			c.Hub.Unregister <- c
			return
		}
	}

}