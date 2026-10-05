package ws

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
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
	userIDParam, uOk := c.Get("userId")
	userID, uValid := userIDParam.(int64)
	if !uOk || !uValid || userID <= 0 {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	usernameParam, nOk := c.Get("username")
	username, nValid := usernameParam.(string)
	if !nOk || !nValid || username == "" {
		helpers.WriteError(c, apierrs.NewUnauthorizedError("unauthorized"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(userID, username, conn, h.hub, h.convService)

	h.hub.Register <- client

	go client.readPump()
	go client.writePump()
	go client.readReceiptWorker()
}
