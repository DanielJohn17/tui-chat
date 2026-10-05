// Package router
package router

import (
	"context"

	"github.com/DanielJohn17/tui-chat/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/middleware"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	User users.UserHandlerInt
	Auth auth.AuthHandlerInt
	Conv conversations.ConvHandlerInt
	WS   ws.WSHandlerInt
}

func NewRouter(h Handlers, ctx context.Context) *gin.Engine {
	if config.ENV.GoEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	_ = router.SetTrustedProxies(nil)

	// Enforce client shared secret validation on all incoming traffic
	router.Use(middleware.ClientSecretAuth(config.ENV.AppSharedSecret))

	// Health check probe
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	subRouter := router.Group("/api/v1")

	// Public auth routes
	rateLimiter := middleware.RateLimiter(ctx, 10, 15)

	subRouter.POST("/auth/register", rateLimiter, h.Auth.RegisterUser)
	subRouter.POST("/auth/login", rateLimiter, h.Auth.LoginUser)

	subRouter.Use(middleware.Auth())
	{
		// users routes
		subRouter.GET("/users/search", middleware.RateLimiter(ctx, 60, 20), h.User.SearchUsers)
		subRouter.GET(
			"/users/:id",
			middleware.ValidateURIParams[types.URLParamInt](),
			h.User.GetUserByID,
		)

		// conversations
		// implement cursor pagination later
		// implement chat history order by latest message update later
		subRouter.GET("/conversations", h.Conv.GetConvsByUserID)
		subRouter.GET("/conversations/bulk", h.Conv.GetBulkChatsByUserID)
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
