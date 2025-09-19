package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
}

func NewRedisLimiter(redisURL string) (*RedisLimiter, error) {
	opt, err := redis.ParseURL(fmt.Sprintf("redis://%s", redisURL))
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisLimiter{client: client}, nil
}

func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int, blockTime time.Duration) (bool, time.Duration, error) {
	blockKey := fmt.Sprintf("block:%s", key)
	countKey := fmt.Sprintf("count:%s", key)

	ttl, err := r.client.TTL(ctx, blockKey).Result()
	if err != nil {
		return false, 0, err
	}

	if ttl > 0 {
		return false, ttl, nil
	}

	currentCount, err := r.client.Get(ctx, countKey).Int()
	if err != nil && err != redis.Nil {
		return false, 0, err
	}

	if currentCount >= limit {
		if err := r.client.Set(ctx, blockKey, "1", blockTime).Err(); err != nil {
			return false, 0, err
		}
		return false, blockTime, nil
	}

	if err := r.client.Incr(ctx, countKey).Err(); err != nil {
		return false, 0, err
	}

	if currentCount == 0 {
		if err := r.client.Expire(ctx, countKey, time.Second).Err(); err != nil {
			return false, 0, err
		}
	}

	return true, 0, nil
}
