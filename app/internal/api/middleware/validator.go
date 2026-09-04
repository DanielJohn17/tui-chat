// Package middleware
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateAndBind[T comparable]() gin.HandlerFunc {
	return func(c *gin.Context) {

		var params T

		if err := c.ShouldBindUri(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Set("params", params)
		c.Next()
	}
}
