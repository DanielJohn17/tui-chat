package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_UnauthenticatedUsesIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := gin.New()
	router.GET("/test", middleware.RateLimiter(ctx, 60, 3), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	// 4th request on the same IP exceeds burst 3
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_AuthenticatedUsersDoNotThrottleEachOther(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		// Mock auth middleware setting userId
		if uid := c.GetHeader("X-Mock-User-ID"); uid != "" {
			if uid == "user1" {
				c.Set("userId", int64(1))
			} else if uid == "user2" {
				c.Set("userId", int64(2))
			}
		}
		c.Next()
	}, middleware.RateLimiter(ctx, 60, 2), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// User 1 uses up their 2 burst tokens on IP 192.0.2.1
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Mock-User-ID", "user1")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	// User 1 is throttled
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Mock-User-ID", "user1")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusTooManyRequests, w.Code)

	// User 2 on the SAME IP still has their 2 burst tokens available!
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Mock-User-ID", "user2")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	// User 2 is now throttled
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Mock-User-ID", "user2")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusTooManyRequests, w.Code)
}
