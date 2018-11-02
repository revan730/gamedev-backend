package src

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/revan730/gamedev-backend/db"
	"github.com/revan730/gamedev-backend/types"
	"go.uber.org/zap"
)

// GameHub contains active game session (client connections)
// and handles client interaction
type GameHub struct {
	clients          map[*Client]bool
	newConnection    chan *Client
	closedConnection chan *Client
	databaseClient   *db.DatabaseClient
	redisClient      *redis.Client
	logger           *zap.Logger
}

func NewGameHub(dbCl *db.DatabaseClient, rCl *redis.Client, logger *zap.Logger) *GameHub {
	return &GameHub{
		clients:          make(map[*Client]bool),
		newConnection:    make(chan *Client),
		closedConnection: make(chan *Client),
		databaseClient:   dbCl,
		logger:           logger,
		redisClient:      rCl,
	}
}

func (g *GameHub) Run() {
	for {
		select {
		case client := <-g.newConnection:
			g.clients[client] = false
			fmt.Println("Client connected")
		case client := <-g.closedConnection:
			fmt.Println("Client disconnected")
			g.SaveUserSession(client.userData)
			delete(g.clients, client)
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

// GetSessionByToken returns session pointer if user's token is valid
// by loading it from database
func (g *GameHub) GetSessionByToken(authToken string) *types.User {
	// Ask redis for user's id by token (if authorized)
	// Load session by users id
	userIdStr, err := g.redisClient.Get(authToken).Result()
	if err != nil {
		return nil
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return nil
	}
	user, err := g.databaseClient.FindUserById(int64(userId))
	if err != nil {
		return nil
	}
	return user
}

func (g *GameHub) SaveUserSession(session *types.User) bool {
	// Save user's session to DB
	err := g.databaseClient.SaveUser(session)
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

func (g *GameHub) GetPageAnswers(pageId int64) []types.Answer {
	answers, err := g.databaseClient.FindPageAnswers(pageId)
	if err != nil {
		g.logError("Unable to get answers", err)
		return nil
	}
	return answers
}
