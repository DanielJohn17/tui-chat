// Package router
package router

import (
	"github.com/DanielJohn17/tui-chat/app/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/app/internal/api/middleware"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	User users.UserHandlerInt
	Auth auth.AuthHandlerInt
	Conv conversations.ConvHandlerInt
}

func NewRouter(h Handlers) *gin.Engine {
	router := gin.Default()

	subRouter := router.Group("/api/v1")

	subRouter.POST("/auth/register", h.Auth.RegisterUser)
	subRouter.POST("/auth/login", h.Auth.LoginUser)

	subRouter.Use(middleware.Auth())
	{
		// users routes
		subRouter.GET(
			"/users/:id",
			middleware.ValidateAndBind[types.URLParamInt](),
			h.User.GetUserByID,
		)

		// conversations
		subRouter.GET("/conversations", h.Conv.GetConvsByUserID)
		subRouter.POST("/conversations", h.Conv.GetOrCreateDirectConversation)
	}
	return router
}
