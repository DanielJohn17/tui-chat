// Package router
package router

import (
	"github.com/DanielJohn17/tui-chat/app/internal/api/middleware"
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	User users.UserHandlerInt
}

func NewRouter(h Handlers) *gin.Engine {
	router := gin.Default()

	subRouter := router.Group("/api/v1")

	subRouter.GET("/users/:id", middleware.ValidateAndBind[types.URLParamInt](), h.User.GetUserByID)
	return router
}
