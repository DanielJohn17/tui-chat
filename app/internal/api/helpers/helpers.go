// Package helpers
package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	apierrors "github.com/DanielJohn17/tui-chat/app/internal/api/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ParseJSON transforms json byte slice into go struct
func ParseJSON[K comparable](c *gin.Context, payload *K) error {
	validate := validator.New()

	if c.Request.Body == nil {
		return apierrors.NewBadRequestError("missing request body")
	}

	if err := json.NewDecoder(c.Request.Body).Decode(payload); err != nil {
		return apierrors.NewBadRequestError(fmt.Sprintf("error decoding payload: %v", err))
	}

	if err := validate.Struct(payload); err != nil {
		return apierrors.NewBadRequestError(
			fmt.Sprintf("validation error: %v", err),
		)
	}

	return nil
}

func WriteJSON[T any](c *gin.Context, code int, data T) {
	c.IndentedJSON(code, gin.H{
		"success": true,
		"data":    data,
	})
}

func WriteError(c *gin.Context, err error) {

	if apiErr, ok := errors.AsType[*apierrors.APIError](err); ok {
		c.AbortWithStatusJSON(apiErr.Code, gin.H{
			"success": false,
			"error":   apiErr.Message,
		})
		return
	}

	// Fallback for unhandled raw errors
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   "internal server error",
	})
}
