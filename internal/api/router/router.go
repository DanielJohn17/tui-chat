// Package router
package router

import (
	"github.com/DanielJohn17/tui-chat/app/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/app/internal/api/middleware"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
	"github.com/DanielJohn17/tui-chat/app/internal/api/ws"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	User users.UserHandlerInt
	Auth auth.AuthHandlerInt
	Conv conversations.ConvHandlerInt
	WS   ws.WSHandlerInt
}

func NewRouter(h Handlers) *gin.Engine {
	if config.ENV.GoEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	_ = router.SetTrustedProxies(nil)

	// Enable CORS & client header validation
	router.Use(middleware.CORS())

	subRouter := router.Group("/api/v1")

	// Public auth routes
	subRouter.POST("/auth/register", h.Auth.RegisterUser)
	subRouter.POST("/auth/login", h.Auth.LoginUser)

	subRouter.Use(middleware.Auth())
	{
		// users routes
		subRouter.GET(
			"/users/:id",
			middleware.ValidateURIParams[types.URLParamInt](),
			h.User.GetUserByID,
		)

		// conversations
		// implement cursor pagination later
		// implement chat history order by latest message update later
		subRouter.GET("/conversations", h.Conv.GetConvsByUserID)
		subRouter.GET(
			"/conversations/:id/chats",
			middleware.ValidateURIParams[types.URLParamInt](),
			middleware.ValidateQueryParams(),
			h.Conv.GetConvChats,
		)
		subRouter.POST("/conversations", h.Conv.GetOrCreateDirectConversation)
		subRouter.DELETE(
			"/conversations/:id",
			middleware.ValidateURIParams[types.URLParamInt](),
			h.Conv.WipeConversation,
		)

		// websocket routes
		subRouter.GET(
			"/ws",
			h.WS.HandleWS,
		)
	}
	return router
}
