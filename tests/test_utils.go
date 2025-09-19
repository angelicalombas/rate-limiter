package tests

import (
	"net/http"
	"net/http/httptest"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/middleware"

	"github.com/gin-gonic/gin"
)

// CreateTestServer cria servidor de teste com configuração personalizada
func CreateTestServer(limiter limiter.RateLimiter, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	middlewareConfig := &middleware.Config{
		RateLimitIP:      cfg.RateLimitIP,
		RateLimitToken:   cfg.RateLimitToken,
		BlockTime:        cfg.BlockTime,
		EnableIPLimit:    cfg.EnableIPLimit,
		EnableTokenLimit: cfg.EnableTokenLimit,
	}

	r.Use(middleware.RateLimitMiddleware(limiter, middlewareConfig))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "Request successful",
			"timestamp": time.Now().Unix(),
		})
	})

	return r
}

// MakeRequest faz uma requisição para teste
func MakeRequest(router *gin.Engine, method, url, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, nil)
	if token != "" {
		req.Header.Set("API_KEY", token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
