package ws

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/app/internal/api/errors"
	"github.com/DanielJohn17/tui-chat/app/internal/api/helpers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader websocket.Upgrader
	goEnv    string
)

func init() {
	goEnv = config.ENV.GoEnv
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if goEnv == "development" {
				return true
			}
			// In production, allow connections from verified TUI client header or non-browser agents
			if r.Header.Get("X-Client-App") == "tui-chat" {
				return true
			}
			origin := r.Header.Get("Origin")
			return origin == ""
		},
	}
}

type WSHandlerInt interface {
	HandleWS(c *gin.Context)
}

type WSHandler struct {
	hub         *Hub
	convService conversations.ConvServiceInt
}

func NewWSHanler(hub *Hub, convService conversations.ConvServiceInt) *WSHandler {
	return &WSHandler{hub: hub, convService: convService}
}

var _ WSHandlerInt = (*WSHandler)(nil)

func (h *WSHandler) HandleWS(c *gin.Context) {
	userIDParam, exists := c.Get("userId")
	if !exists {
		helpers.WriteError(c, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	usernameParam, exists := c.Get("username")
	if !exists {
		helpers.WriteError(c, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	userID, ok := userIDParam.(int64)
	if !ok {
		helpers.WriteError(c, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	username, ok := usernameParam.(string)
	if !ok {
		helpers.WriteError(c, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		UserID:      userID,
		Username:    username,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		Hub:         h.hub,
		convService: h.convService,
	}

	client.ActiveConvID.Store(0)

	h.hub.Register <- client

	go client.readPump()
	go client.writePump()
}
