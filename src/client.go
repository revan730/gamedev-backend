package src

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/revan730/gamedev-backend/types"
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
	c.SendCurrentPage()
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

// TODO: And this one too
func (c *Client) SendCurrentPage() {
	page := c.hub.GetPage(c.session.userData.CurrentPage)
	c.sendJSON(page)
}

// TODO: Using reflect?
func (c *Client) RecalculateStats(answer *types.Answer) {
	c.session.userData.Knowledge += answer.Knowledge
	c.session.userData.Performance += answer.Performance
	c.session.userData.Sober += answer.Sober
	c.session.userData.Prestige += answer.Prestige
	c.session.userData.Connections += answer.Connections
}

// NextPage proceeds game session to next page
// handles questions and jump logic
// TODO: Page struct should contain array of answers for quick access
func (c *Client) NextPage(jsonMap map[string]interface{}) error {
	currentPage := c.hub.GetPage(c.session.userData.CurrentPage)
	// Check if current page has questions
	// and handle them
	if currentPage.IsQuestion == true {
		// Load answer
		answerId, ok := jsonMap["answerId"].(int64)
		if ok == false {
			return errors.New("NextPage: bad or missing answerId")
		}
		answer := c.hub.GetAnswer(answerId)
		if answer == nil {
			return errors.New("NextPage: answer not found")
		}
		// Recalculate user stats and set flags according to
		// answer values
		// TODO: Flags
		c.RecalculateStats(answer)
	}
	if currentPage.IsJumper == true {
		// TODO: Jumper logic handle
		return nil
	} else {
		// Linear transition
		c.session.userData.CurrentPage = currentPage.NextPage
		return nil
	}
}

func (c *Client) HandleStoryMessages(jsonMap map[string]interface{}) {
	responseMap := map[string]interface{}{
		"channel":  "story",
		"response": false,
	}
	switch jsonMap["method"] {
	case "save":
		// Save user's progress
		responseMap["response"] = c.hub.SaveUserSession(c.session)
		c.sendJSON(responseMap)
	case "forward":
		err := c.NextPage(jsonMap)
		if err != nil {
			fmt.Println("Failed to go to next page: ", err)
			c.sendJSON(responseMap)
			return
		}
		c.SendCurrentPage()
	default:
		// Unknown method
		c.sendJSON(responseMap)
	}
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
	case "story":
		c.HandleStoryMessages(jsonMap)
	default:
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
