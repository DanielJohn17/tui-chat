package middleware

import (
	"net/http"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/types"
	"github.com/gin-gonic/gin"
)

// ValidateURIParams validates and binds URI path parameters into type T and sets "params" in context.
func ValidateURIParams[T any]() gin.HandlerFunc {
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

// ValidateAndBind is a generic URI validator function for backward compatibility.
func ValidateAndBind[T any]() gin.HandlerFunc {
	return ValidateURIParams[T]()
}

// ValidateQueryParams validates and binds query parameters into types.URLQueryParams and sets "queries" in context.
func ValidateQueryParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		var queryParams types.URLQueryParams
		if err := c.ShouldBindQuery(&queryParams); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if queryParams.Cursor != "" {
			cursorTime, cursorID, err := ParseCursor(queryParams.Cursor)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			queryParams.CursorTime = cursorTime
			queryParams.CursorID = cursorID
		}

		if queryParams.Limit <= 0 {
			queryParams.Limit = 30
		}

		c.Set("queries", queryParams)
		c.Next()
	}
}

// ParseCursor parses a cursor string (Base64 URL-safe or raw) into time.Time and int64.
func ParseCursor(cursor string) (time.Time, int64, error) {
	return helpers.DecodeCursor(cursor)
}
