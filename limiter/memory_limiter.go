package limiter

import (
	"context"
	"sync"
	"time"
)

// MemoryLimiter para testes sem dependência do Redis
type MemoryLimiter struct {
	mu          sync.RWMutex
	counts      map[string]int
	blocks      map[string]time.Time
	expiries    map[string]time.Time
	windowStart map[string]time.Time
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		counts:      make(map[string]int),
		blocks:      make(map[string]time.Time),
		expiries:    make(map[string]time.Time),
		windowStart: make(map[string]time.Time),
	}
}

func (m *MemoryLimiter) Allow(ctx context.Context, key string, limit int, blockTime time.Duration) (bool, time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	if blockUntil, exists := m.blocks[key]; exists {
		if now.Before(blockUntil) {
			return false, blockUntil.Sub(now), nil
		}
		delete(m.blocks, key)
		delete(m.counts, key)
		delete(m.expiries, key)
		delete(m.windowStart, key)
	}
	if windowStart, exists := m.windowStart[key]; exists {
		if now.Sub(windowStart) >= time.Second {
			// Janela expirada - reseta contador
			delete(m.counts, key)
			delete(m.expiries, key)
			delete(m.windowStart, key)
		}
	}
	if _, exists := m.windowStart[key]; !exists {
		m.windowStart[key] = now
		m.counts[key] = 0
	}
	if m.counts[key] >= limit {
		// Bloqueia por blockTime
		blockUntil := now.Add(blockTime)
		m.blocks[key] = blockUntil
		return false, blockTime, nil
	}

	m.counts[key]++

	m.expiries[key] = m.windowStart[key].Add(time.Second)

	return true, 0, nil
}

// Método para testes - limpa o estado
func (m *MemoryLimiter) TestCleanup(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.counts, key)
	delete(m.blocks, key)
	delete(m.expiries, key)
	delete(m.windowStart, key)

	return nil
}

// Método para forçar limpeza de todos os estados expirados
func (m *MemoryLimiter) ForceCleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for key, blockUntil := range m.blocks {
		if now.After(blockUntil) {
			delete(m.blocks, key)
			delete(m.counts, key)
			delete(m.expiries, key)
			delete(m.windowStart, key)
		}
	}

	for key, expiry := range m.expiries {
		if now.After(expiry) {
			delete(m.counts, key)
			delete(m.expiries, key)
			delete(m.windowStart, key)
		}
	}
}
