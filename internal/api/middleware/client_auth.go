// Package middleware provides HTTP middlewares for routing, authentication, and security verification.
package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/gin-gonic/gin"
)

// ClientSecretAuth validates that incoming requests originate from an authorized client.
// It checks the 'X-App-Secret' (or 'X-Client-Secret') header using constant-time SHA-256 hash comparison.
func ClientSecretAuth(expectedSecret string) gin.HandlerFunc {
	if expectedSecret == "" {
		panic("ClientSecretAuth: expectedSecret cannot be empty")
	}

	expectedHash := sha256.Sum256([]byte(expectedSecret))

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Allow unauthenticated health probes
		if path == "/health" || path == "/healthz" || path == "/livez" || path == "/readyz" {
			c.Next()
			return
		}

		providedSecret := c.GetHeader(config.ClientSecretHeader)
		if providedSecret == "" {
			providedSecret = c.GetHeader(config.AltClientSecretHeader)
		}

		if providedSecret == "" {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"success": false,
					"error":   "Forbidden: missing client secret header",
				},
			)
			return
		}

		providedHash := sha256.Sum256([]byte(providedSecret))
		if subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) != 1 {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"success": false,
					"error":   "Forbidden: invalid client secret",
				},
			)
			return
		}

		c.Next()
	}
}
