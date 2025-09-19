package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
)

func TestIPRateLimiting(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()
	cfg := &config.Config{
		RateLimitIP:      5,
		RateLimitToken:   100,
		BlockTime:        2 * time.Second,
		EnableIPLimit:    true,
		EnableTokenLimit: true,
	}

	router := CreateTestServer(memoryLimiter, cfg)

	// Testa 6 requisições do mesmo IP (limite: 5)
	for i := 0; i < 6; i++ {
		w := MakeRequest(router, "GET", "/", "")

		if i < 5 {
			if w.Code != http.StatusOK {
				t.Errorf("Request %d should be allowed, got status %d", i+1, w.Code)
			}
		} else {
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Request %d should be blocked, got status %d", i+1, w.Code)
			}

			// Verificação simples mas eficaz
			if w.Body.String() == "" {
				t.Error("Response body should not be empty")
			}
		}
	}
}

func TestTokenRateLimiting(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()
	cfg := &config.Config{
		RateLimitIP:      10,
		RateLimitToken:   3,
		BlockTime:        1 * time.Second,
		EnableIPLimit:    true,
		EnableTokenLimit: true,
	}

	router := CreateTestServer(memoryLimiter, cfg)
	token := "test-token-123"

	// Testa 4 requisições com o mesmo token (limite: 3)
	for i := 0; i < 4; i++ {
		w := MakeRequest(router, "GET", "/", token)

		if i < 3 {
			if w.Code != http.StatusOK {
				t.Errorf("Token request %d should be allowed, got status %d", i+1, w.Code)
			}
		} else {
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Token request %d should be blocked, got status %d", i+1, w.Code)
			}
		}
	}
}

func TestTokenPrecedenceOverIP(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()
	cfg := &config.Config{
		RateLimitIP:      2,  // Baixo limite para IP
		RateLimitToken:   10, // Alto limite para token
		BlockTime:        1 * time.Second,
		EnableIPLimit:    true,
		EnableTokenLimit: true,
	}

	router := CreateTestServer(memoryLimiter, cfg)
	token := "precedence-token"

	// Faz 5 requisições com token (deveriam passar devido ao limite alto do token)
	for i := 0; i < 5; i++ {
		w := MakeRequest(router, "GET", "/", token)

		if w.Code != http.StatusOK {
			t.Errorf("Request with token %d should be allowed (token precedence), got status %d", i+1, w.Code)
		}
	}
}

func TestBlockTime(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()

	// Testa diretamente o limiter sem o middleware
	ctx := context.Background()
	key := "test-key"
	limit := 1
	blockTime := 100 * time.Millisecond

	// Primeira requisição - deve passar
	allowed1, _, err := memoryLimiter.Allow(ctx, key, limit, blockTime)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !allowed1 {
		t.Error("First request should be allowed")
	}

	// Segunda requisição - deve ser bloqueada
	allowed2, ttl, err := memoryLimiter.Allow(ctx, key, limit, blockTime)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if allowed2 {
		t.Error("Second request should be blocked")
	}
	if ttl > blockTime || ttl <= 0 {
		t.Errorf("TTL should be between 0 and %v, got %v", blockTime, ttl)
	}

	// Espera o bloqueio expirar
	time.Sleep(blockTime + 50*time.Millisecond)

	// Terceira requisição - deve passar após bloqueio
	allowed3, _, err := memoryLimiter.Allow(ctx, key, limit, blockTime)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if !allowed3 {
		t.Error("Third request should be allowed after block time")
	}
}
func TestDisabledLimiters(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()

	// Testa com IP limiter desabilitado
	cfgIPDisabled := &config.Config{
		RateLimitIP:      1,
		RateLimitToken:   1,
		BlockTime:        1 * time.Second,
		EnableIPLimit:    false, // IP limiter desabilitado
		EnableTokenLimit: true,
	}

	router := CreateTestServer(memoryLimiter, cfgIPDisabled)

	// Várias requisições sem token - devem passar (IP limiter desabilitado)
	for i := 0; i < 5; i++ {
		w := MakeRequest(router, "GET", "/", "")
		if w.Code != http.StatusOK {
			t.Errorf("Request %d should be allowed (IP limiter disabled), got status %d", i+1, w.Code)
		}
	}

	// Testa com Token limiter desabilitado
	cfgTokenDisabled := &config.Config{
		RateLimitIP:      5,
		RateLimitToken:   1,
		BlockTime:        1 * time.Second,
		EnableIPLimit:    true,
		EnableTokenLimit: false, // Token limiter desabilitado
	}

	router2 := CreateTestServer(memoryLimiter, cfgTokenDisabled)

	// Várias requisições com token - devem passar (Token limiter desabilitado)
	for i := 0; i < 5; i++ {
		w := MakeRequest(router2, "GET", "/", "any-token")
		if w.Code != http.StatusOK {
			t.Errorf("Request with token %d should be allowed (token limiter disabled), got status %d", i+1, w.Code)
		}
	}
}

func TestDifferentTokensDifferentCounters(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()
	cfg := &config.Config{
		RateLimitIP:      10,
		RateLimitToken:   2,
		BlockTime:        1 * time.Second,
		EnableIPLimit:    true,
		EnableTokenLimit: true,
	}

	router := CreateTestServer(memoryLimiter, cfg)

	// Testa com token 1
	for i := 0; i < 3; i++ {
		w := MakeRequest(router, "GET", "/", "token-1")
		if i < 2 {
			if w.Code != http.StatusOK {
				t.Errorf("Token 1 request %d should be allowed, got status %d", i+1, w.Code)
			}
		} else {
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Token 1 request %d should be blocked, got status %d", i+1, w.Code)
			}
		}
	}

	// Testa com token 2 (deve ter contador separado)
	for i := 0; i < 2; i++ {
		w := MakeRequest(router, "GET", "/", "token-2")
		if w.Code != http.StatusOK {
			t.Errorf("Token 2 request %d should be allowed (separate counter), got status %d", i+1, w.Code)
		}
	}
}

func TestConcurrentRequests(t *testing.T) {
	memoryLimiter := limiter.NewMemoryLimiter()
	cfg := &config.Config{
		RateLimitIP:      10,
		RateLimitToken:   10,
		BlockTime:        1 * time.Second,
		EnableIPLimit:    true,
		EnableTokenLimit: true,
	}

	router := CreateTestServer(memoryLimiter, cfg)

	// Canal para resultados
	results := make(chan int, 15)

	// Faz 15 requisições concorrentes
	for i := 0; i < 15; i++ {
		go func() {
			w := MakeRequest(router, "GET", "/", "")
			results <- w.Code
		}()
	}

	// Coleta resultados
	successCount := 0
	blockedCount := 0
	for i := 0; i < 15; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		} else if code == http.StatusTooManyRequests {
			blockedCount++
		}
	}

	// Devem ter 10 sucessos e 5 bloqueios
	if successCount != 10 {
		t.Errorf("Expected 10 successful requests, got %d", successCount)
	}
	if blockedCount != 5 {
		t.Errorf("Expected 5 blocked requests, got %d", blockedCount)
	}
}

func TestRedisLimiterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	// Teste de integração com Redis real
	redisLimiter, err := limiter.NewRedisLimiter("localhost:6379")
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Testa diretamente o limiter
	ctx := context.Background()
	key := "test-redis-key"

	// Testa 4 requisições
	for i := 0; i < 4; i++ {
		allowed, _, err := redisLimiter.Allow(ctx, key, 3, 1*time.Second)
		if err != nil {
			t.Fatalf("Redis error: %v", err)
		}

		if i < 3 && !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
		if i == 3 && allowed {
			t.Errorf("Request %d should be blocked", i+1)
		}
	}

}
