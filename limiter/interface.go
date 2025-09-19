package limiter

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, blockTime time.Duration) (bool, time.Duration, error)
}
