package src

import (
	"encoding/json"
	"fmt"
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
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	conn    *websocket.Conn
	hub     *GameHub
	session *Session
	send    chan []byte
}

func (c *Client) Authorize(authToken string) {
	// Get client's session from hub
	session := c.hub.FindSessionByToken(authToken)
	responseMap := map[string]interface{}{
		"channel":  "auth",
		"response": false,
	}
	if session == nil {
		// Write message informing that authorization failed
		c.sendJSON(responseMap)
		return
	}
	// Inform user that authorization was successfull
	// And send session data
	c.session = session
	c.SendSessionInfo()
	// TODO: Send session data
}

func (c *Client) sendJSON(d interface{}) {
	j, _ := json.Marshal(d)
	c.send <- j
}

// TODO: Very likely to be changed
func (c *Client) SendSessionInfo() {
	c.sendJSON(c.session.userData)
}

func (c *Client) HandleClientMessage(jsonInt interface{}) {
	jsonMap := jsonInt.(map[string]interface{})
	switch jsonMap["channel"] {
	case "auth":
		responseMap := map[string]interface{}{
			"channel":  "auth",
			"response": false,
		}
		token, ok := jsonMap["authToken"].(string)
		if ok != true {
			c.sendJSON(responseMap)
			return
		}
		c.Authorize(token)
	default:
		fmt.Println("fuck wat")
		c.sendJSON(map[string]string{"err": "wtf"})
	}
}

func (c *Client) Reader() {
	defer func() {
		c.hub.closedConnection <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var msg interface{}
	for {
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}
		// TODO: Handle messages from client somewhere right here
		fmt.Printf("Message: %s", msg)
		c.HandleClientMessage(msg)
	}
}

func (c *Client) Writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
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
			w.Write(message)

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
