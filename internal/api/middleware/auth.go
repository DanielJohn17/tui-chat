// Package middleware
package middleware

import (
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "Missing authorization header"},
			)
			return
		}

		const prefix = "Bearer "
		if len(tokenString) < len(prefix) || tokenString[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "Invalid authorization format, expected 'Bearer <token>'"},
			)
			return
		}

		tokenString = tokenString[len(prefix):]

		userToken, err := helpers.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userId", userToken.ID)
		c.Set("username", userToken.Username)

		c.Next()
	}
}
