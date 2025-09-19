package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"rate-limiter/limiter"

	"github.com/gin-gonic/gin"
)

type Config struct {
	RateLimitIP      int
	RateLimitToken   int
	BlockTime        time.Duration
	EnableIPLimit    bool
	EnableTokenLimit bool
}

func RateLimitMiddleware(limiter limiter.RateLimiter, config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		var identifier string
		var limit int

		// Verifica token primeiro (tem precedência sobre IP)
		token := c.GetHeader("API_KEY")
		if token != "" && config.EnableTokenLimit {
			identifier = "token:" + token
			limit = config.RateLimitToken
		} else if config.EnableIPLimit {
			identifier = "ip:" + getClientIP(c)
			limit = config.RateLimitIP
		} else {
			c.Next()
			return
		}

		allowed, ttl, err := limiter.Allow(ctx, identifier, limit, config.BlockTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "you have reached the maximum number of requests or actions allowed within a certain time frame",
				"retry_after": ttl.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
