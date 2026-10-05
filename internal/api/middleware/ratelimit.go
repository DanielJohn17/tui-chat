package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenLimiter struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter creates a token-bucket rate limiter that keys by (userID + IP).
// ctx: context with cancel
// ratePerMinute: number of tokens replenished per minute (e.g. 5)
// burst: maximum burst capacity (e.g. 10)
func RateLimiter(ctx context.Context, ratePerMinute float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]*tokenLimiter)

	// prevents a panic on <-ctx.Done() if a test ever calls the router or middleware with a nil context.
	if ctx == nil {
		ctx = context.Background()
	}

	startCleaner(ctx, &mu, clients)

	ratePerSec := ratePerMinute / 60.0
	return func(c *gin.Context) {
		key := resolveClientKey(c)
		now := time.Now()

		mu.Lock()
		lim, exists := clients[key]
		if !exists {
			lim = &tokenLimiter{
				tokens:     float64(burst) - 1.0,
				lastRefill: now,
			}
			clients[key] = lim
			mu.Unlock()
			c.Next()
			return
		}

		// refill tokens based on elapsed time
		elapsed := now.Sub(lim.lastRefill).Seconds()
		lim.tokens += elapsed * ratePerSec
		if lim.tokens > float64(burst) {
			lim.tokens = float64(burst)
		}
		lim.lastRefill = now

		if lim.tokens < 1.0 {
			mu.Unlock()
			c.Header("Retry-After", "15")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests. Please slow down and try again.",
			})

			return
		}

		lim.tokens -= 1.0
		mu.Unlock()
		c.Next()
	}
}

func startCleaner(ctx context.Context, mu *sync.Mutex, clients map[string]*tokenLimiter) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				func() {
					mu.Lock()
					defer mu.Unlock()

					now := time.Now()
					for key, lim := range clients {
						if now.Sub(lim.lastRefill) > 10*time.Minute {
							delete(clients, key)
						}
					}
				}()
			}
		}
	}()
}

func resolveClientKey(c *gin.Context) string {
	userID, ok := c.Get("userId")
	id, valid := userID.(int64)
	if ok && valid && id > 0 {
		return fmt.Sprintf("user:%d:%s", id, c.ClientIP())
	}
	return fmt.Sprintf("ip:%s", c.ClientIP())
}
