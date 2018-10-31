package src

import (
	"fmt"
	"net/http"

	"github.com/revan730/gamedev-backend/db"
	"github.com/revan730/gamedev-backend/types"
	"go.uber.org/zap"
)

// GameHub contains active game session (client connections)
// and handles client interaction
type GameHub struct {
	clients          map[*Client]bool
	sessions         map[string]*Session
	newConnection    chan *Client
	closedConnection chan *Client
	newSession       chan *Session
	databaseClient   *db.DatabaseClient
	logger           *zap.Logger
}

func NewGameHub(dbCl *db.DatabaseClient, logger *zap.Logger) *GameHub {
	return &GameHub{
		sessions:         make(map[string]*Session),
		clients:          make(map[*Client]bool),
		newConnection:    make(chan *Client),
		closedConnection: make(chan *Client),
		newSession:       make(chan *Session),
		databaseClient:   dbCl,
		logger:           logger,
	}
}

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
		g.logError("Unable to start WS server", err)
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

func (g *GameHub) logError(msg string, err error) {
	defer g.logger.Sync()
	g.logger.Error(msg, zap.String("packageLevel", "hub"), zap.Error(err))
}

func (g *GameHub) logInfo(msg string) {
	defer g.logger.Sync()
	g.logger.Info("INFO", zap.String("msg", msg), zap.String("packageLevel", "hub"))
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

func (g *GameHub) SaveUserSession(session *Session) bool {
	// Save user's session to DB
	err := g.databaseClient.SaveUser(session.userData)
	if err != nil {
		g.logError("Unable to save user's session", err)
		return false
	}
	return true
}

func (g *GameHub) GetPage(pageId int64) *types.Page {
	page, err := g.databaseClient.FindPageById(pageId)
	if err != nil {
		g.logError("Unable to get page", err)
		return nil
	}
	return page
}

func (g *GameHub) GetAnswer(answerId int64) *types.Answer {
	answer, err := g.databaseClient.FindAnswerById(answerId)
	if err != nil {
		g.logError("Unable to get answer", err)
		return nil
	}
	return answer
}
