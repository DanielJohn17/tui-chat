// Package helpers
package helpers

import (
	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/gin-gonic/gin"
)

func WriteJSON[T comparable](c *gin.Context, response types.APIResponse[T]) {
	c.IndentedJSON(response.Status, gin.H{
		"success": response.Success,
		"data":    response.Data,
	})
}

func WriteListJSON[T comparable](c *gin.Context, response types.APIListResponse[T]) {
	c.IndentedJSON(response.Status, gin.H{
		"success": response.Success,
		"data":    response.Data,
		"meta":    response.Meta,
	})
}

func WriteError(c *gin.Context, apiErr types.APIErrorResponse) {
	c.AbortWithStatusJSON(apiErr.Status, gin.H{
		"success": apiErr.Success,
		"error":   apiErr.Message,
	})
}
