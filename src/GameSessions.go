package src

import (
	"fmt"
	"net/http"

	"github.com/revan730/gamedev-backend/types"
)

// GameHub contains active game session (client connections)
// and handles client interaction
type GameHub struct {
	clients          map[*Client]bool
	sessions         map[string]*Session
	newConnection    chan *Client
	closedConnection chan *Client
	newSession       chan *Session
}

func NewGameHub() *GameHub {
	return &GameHub{
		sessions:         make(map[string]*Session),
		clients:          make(map[*Client]bool),
		newConnection:    make(chan *Client),
		closedConnection: make(chan *Client),
		newSession:       make(chan *Session),
	}
}

// TODO: Db connection
// TODO: End session handling
func (g *GameHub) Run() {
	for {
		select {
		case client := <-g.newConnection:
			g.clients[client] = false
			fmt.Println("Client connected")
		case client := <-g.closedConnection:
			fmt.Println("Client disconnected")
			delete(g.clients, client)
		case session := <-g.newSession:
			fmt.Println("New session, user login " + session.userData.Login)
			g.sessions[session.authToken] = session
		}
	}
}

// serveWs handles websocket requests from the peer.
func (g *GameHub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &Client{hub: g, conn: conn, send: make(chan []byte, 256)}
	client.hub.newConnection <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.Writer()
	go client.Reader()
}

type Session struct {
	userData  *types.User
	authToken string
}

// FindSessionByUser returns session pointer if session of user
// with provided id exists
func (g *GameHub) FindSessionByUser(userId int64) *Session {
	for _, s := range g.sessions {
		if s.userData.Id == userId {
			return s
		}
	}
	return nil
}

// FindSessionByToken returns session pointer if session of user
// with provided authToken exists
func (g *GameHub) FindSessionByToken(authToken string) *Session {
	for _, s := range g.sessions {
		if s.authToken == authToken {
			return s
		}
	}
	return nil
}
