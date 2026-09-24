// Package helpers
package helpers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	apierrors "github.com/DanielJohn17/tui-chat/internal/api/errors"
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

// WriteJSON writes indented JSON response
func WriteJSON[T any](c *gin.Context, code int, data T) {
	c.IndentedJSON(code, gin.H{
		"success": true,
		"data":    data,
	})
}

// WriteJSONWithMeta writes indented JSON response with pagination metadata
func WriteJSONWithMeta[T any](c *gin.Context, code int, data T, meta any) {
	resp := gin.H{
		"success": true,
		"data":    data,
	}
	if meta != nil {
		resp["meta"] = meta
	}
	c.IndentedJSON(code, resp)
}

// WriteError writes error JSON response
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

// EncodeCursor encodes a timestamp string and ID into a URL-safe Base64 cursor string.
func EncodeCursor(timeStr string, id int64) string {
	payload := fmt.Sprintf("%s,%d", timeStr, id)
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

// DecodeCursor decodes a URL-safe Base64 cursor (or raw formatted string) into time.Time and int64.
func DecodeCursor(cursor string) (time.Time, int64, error) {
	if cursor == "" {
		return time.Time{}, 0, errors.New("empty cursor")
	}

	decodedBytes, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		decodedBytes, err = base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			decodedBytes, err = base64.StdEncoding.DecodeString(cursor)
			if err != nil {
				// Fallback to unencoded raw string
				decodedBytes = []byte(cursor)
			}
		}
	}

	payload := string(decodedBytes)

	// Support delimiters: ',', '_', '+', ' '
	idx := strings.LastIndex(payload, ",")
	if idx == -1 {
		idx = strings.LastIndex(payload, "_")
	}
	if idx == -1 {
		idx = strings.LastIndex(payload, "+")
	}
	if idx == -1 {
		idx = strings.LastIndex(payload, " ")
	}

	if idx == -1 || idx == 0 || idx == len(payload)-1 {
		return time.Time{}, 0, fmt.Errorf("invalid cursor format")
	}

	timeStr := payload[:idx]
	idStr := payload[idx+1:]

	// If '+' in timezone offset was decoded to a space by the URL query parser, restore '+'
	if strings.Contains(timeStr, "T") && strings.Contains(timeStr, " ") {
		lastSpace := strings.LastIndex(timeStr, " ")
		timeStr = timeStr[:lastSpace] + "+" + timeStr[lastSpace+1:]
	}

	cursorTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		cursorTime, err = time.Parse(time.RFC3339Nano, timeStr)
		if err != nil {
			return time.Time{}, 0, fmt.Errorf("invalid cursor time format: %w", err)
		}
	}

	cursorID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || cursorID <= 0 {
		return time.Time{}, 0, fmt.Errorf("invalid cursor id, must be a positive integer")
	}

	return cursorTime, cursorID, nil
}
