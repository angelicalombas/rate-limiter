package main

import (
	"log"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.LoadConfig()

	rateLimiter, err := limiter.NewRedisLimiter(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	middlewareConfig := &middleware.Config{
		RateLimitIP:      cfg.RateLimitIP,
		RateLimitToken:   cfg.RateLimitToken,
		BlockTime:        cfg.BlockTime,
		EnableIPLimit:    cfg.EnableIPLimit,
		EnableTokenLimit: cfg.EnableTokenLimit,
	}

	r := gin.Default()

	r.Use(middleware.RateLimitMiddleware(rateLimiter, middlewareConfig))

	r.Any("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "Request successful",
			"timestamp": time.Now().Unix(),
		})
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
