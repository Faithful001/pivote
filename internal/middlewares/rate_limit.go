package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	redisClient "pivote/internal/infra/redis"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	Max int
	Window time.Duration
	KeyPrefix string
	UseUserID bool
}

func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Build the rate-limit key
		identifier := c.ClientIP()
		if cfg.UseUserID {
			if u, err := GetUser(c); err == nil {
				identifier = u.ID.String()
			}
		}
		key := fmt.Sprintf("rl:%s:%s", cfg.KeyPrefix, identifier)

		pipe := redisClient.Client.Pipeline()
		incr := pipe.Incr(ctx, key)

		
		if _, err := pipe.Exec(ctx); err != nil {
			c.Next()
			return
		}

		count := incr.Val()

		if count == 1 {
			redisClient.Client.Expire(ctx, key, cfg.Window)
		}

		// Set informational headers
		remaining := int64(cfg.Max) - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", cfg.Window.String())

		if count > int64(cfg.Max) {
			// Calculate TTL for Retry-After
			ttl, err := redisClient.Client.TTL(ctx, key).Result()
			if err != nil || ttl == -1*time.Second { // if there's an error or a ttl was not set when pipe.Expire was called
				ttl = cfg.Window
			}
			c.Header("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"statusCode": http.StatusTooManyRequests,
				"success":    false,
				"message":    "Too many requests. Please try again later.",
				"data":       nil,
			})
			return
		}

		c.Next()
	}
}

func RateLimitByIP(prefix string, max int, window time.Duration) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		KeyPrefix: prefix,
		Max:       max,
		Window:    window,
		UseUserID: false,
	})
}

func lookupTTL(ctx context.Context, key string) (time.Duration, error) {
	return redisClient.Client.TTL(ctx, key).Result()
}

// need in future extensions of this file.
var _ = redis.Nil
