package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipLimiter struct {
	tokens     float64
	lastRefill time.Time
}

// AuthRateLimiter creates an IP-based token-bucket rate limiter.
// ctx: context with cancel
// ratePerMinute: number of tokens replenished per minute (e.g. 5)
// burst: maximum burst capacity (e.g. 10)
func AuthRateLimiter(ctx context.Context, ratePerMinute float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]*ipLimiter)

	// prevents a panic on <-ctx.Done() if a test ever calls the router or middleware with a nil context.
	if ctx == nil {
		ctx = context.Background()
	}

	startCleaner(ctx, &mu, clients)

	ratePerSec := ratePerMinute / 60.0
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		lim, exists := clients[ip]
		if !exists {
			lim = &ipLimiter{
				tokens:     float64(burst) - 1.0,
				lastRefill: now,
			}
			clients[ip] = lim
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
				"error":   "Too many unauthenticated attempts. Please slow down and try again.",
			})

			return
		}

		lim.tokens -= 1.0
		mu.Unlock()
		c.Next()
	}
}

func startCleaner(ctx context.Context, mu *sync.Mutex, clients map[string]*ipLimiter) {
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
					for ip, lim := range clients {
						if now.Sub(lim.lastRefill) > 10*time.Minute {
							delete(clients, ip)
						}
					}
				}()
			}
		}
	}()
}
