// Package helpers
package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ParseJSON transforms json byte slice into go struct
func ParseJSON[K comparable](c *gin.Context, payload *K) *types.APIErrorResponse {
	validate := validator.New()

	if c.Request.Body == nil {
		return &types.APIErrorResponse{
			Status:  http.StatusBadRequest,
			Success: false,
			Message: "missing request body",
		}
	}

	if err := json.NewDecoder(c.Request.Body).Decode(payload); err != nil {
		return &types.APIErrorResponse{
			Status:  http.StatusBadRequest,
			Success: false,
			Message: fmt.Errorf("error decoding payload: %w", err).Error(),
		}
	}

	if err := validate.Struct(payload); err != nil {
		return &types.APIErrorResponse{
			Status:  http.StatusBadRequest,
			Success: false,
			Message: fmt.Errorf("validation error: %w", err).Error(),
		}
	}

	return nil
}

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
