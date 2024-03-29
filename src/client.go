package src

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/revan730/gamedev-backend/lua"
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
	conn     *websocket.Conn
	hub      *GameHub
	userData *types.User
	send     chan interface{}
}

func (c *Client) Authorize(authToken string) {
	// Get client's session from db via hub
	session := c.hub.GetSessionByToken(authToken)
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
	c.userData = session
	c.SendSessionInfo()
	c.SendCurrentPage()
}

func (c *Client) sendJSON(d interface{}) {
	//j, _ := json.Marshal(d)
	c.send <- d
}

// TODO: Very likely to be changed
func (c *Client) SendSessionInfo() {
	jsonMap := map[string]interface{}{
		"channel": "stats",
	}
	jsonMap["stats"] = c.userData
	c.sendJSON(jsonMap)
}

func (c *Client) SendCurrentPage() {
	jsonMap := map[string]interface{}{
		"channel": "story_text",
	}
	page := c.hub.GetPage(c.userData.CurrentPage)
	jsonMap["text"] = page.Text
	if page.IsQuestion == true {
		answers := c.hub.GetPageAnswers(page.Id)
		jsonMap["answers"] = answers
	}
	c.sendJSON(jsonMap)
}

// TODO: Using reflect?
func (c *Client) recalculateStats(answer *types.Answer) {
	c.userData.Knowledge += answer.Knowledge
	c.userData.Performance += answer.Performance
	c.userData.Sober += answer.Sober
	c.userData.Prestige += answer.Prestige
	c.userData.Connections += answer.Connections
	c.userData.Praepostor += answer.Praepostor
}

// NextPage proceeds game session to next page
// handles questions and jump logic
func (c *Client) NextPage(jsonMap map[string]interface{}) error {
	currentPage := c.hub.GetPage(c.userData.CurrentPage)
	// Check if current page has questions
	// and handle them
	if currentPage.IsQuestion == true {
		// Load answer
		answerId, ok := jsonMap["answerId"].(float64)
		if ok == false {
			return errors.New("NextPage: bad or missing answerId")
		}
		answer := c.hub.GetAnswer(int64(answerId))
		if answer == nil {
			return errors.New("NextPage: answer not found")
		}
		// Recalculate user stats and set flags according to
		// answer values
		c.recalculateStats(answer)
		c.userData.MergeFlags(answer.Flags)
	}
	if currentPage.IsJumper == true {
		interpreter := lua.NewInterpreter(c.userData)
		interpreter.DoString(currentPage.JumperLogic)
		return nil
	} else {
		// If next page is null here, story has come to end
		// Restart from first page and reset stats and flags(?)
		if currentPage.NextPage == 0 {
			c.ResetStory()
		} else {
			// Linear transition
			c.userData.CurrentPage = currentPage.NextPage
		}
		return nil
	}
}

func (c *Client) ResetStory() {
	c.userData.Reset()
	c.SendSessionInfo()
}

func (c *Client) HandleStoryMessages(jsonMap map[string]interface{}) {
	responseMap := map[string]interface{}{
		"channel":  "story",
		"response": false,
	}
	switch jsonMap["channel"] {
	case "story_save":
		// Save user's progress
		responseMap["response"] = c.hub.SaveUserSession(c.userData)
		responseMap["channel"] = "story_save"
		c.sendJSON(responseMap)
	case "story_move":
		err := c.NextPage(jsonMap)
		if err != nil {
			fmt.Println("Failed to go to next page: ", err)
			responseMap["channel"] = "story_move"
			c.sendJSON(responseMap)
			return
		}
		c.SendSessionInfo()
		c.SendCurrentPage()
	case "story_reset":
		c.ResetStory()
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
	case "story_move":
		c.HandleStoryMessages(jsonMap)
	case "story_save":
		c.HandleStoryMessages(jsonMap)
	case "story_reset":
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

			c.conn.WriteJSON(message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
