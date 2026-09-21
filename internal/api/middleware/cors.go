// Package middleware provides HTTP middlewares for routing, authentication, validation, and CORS.
package middleware

import (
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var goEnv string

func init() {
	goEnv = config.ENV.GoEnv
}

// CORS returns the standard gin-contrib/cors middleware configured for development and production environments.
func CORS() gin.HandlerFunc {
	if goEnv == "development" {
		cfg := cors.DefaultConfig()
		cfg.AllowAllOrigins = true
		cfg.AllowHeaders = append(cfg.AllowHeaders, "Authorization", "X-Client-App")
		return cors.New(cfg)
	}

	// Production CORS configuration
	cfg := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Client-App"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(cfg)
}
