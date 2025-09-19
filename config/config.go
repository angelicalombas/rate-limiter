package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisURL         string
	RateLimitIP      int
	RateLimitToken   int
	BlockTime        time.Duration
	EnableIPLimit    bool
	EnableTokenLimit bool
}

func LoadConfig() *Config {
	return &Config{
		RedisURL:         getEnv("REDIS_URL", "localhost:6379"),
		RateLimitIP:      getEnvAsInt("RATE_LIMIT_IP", 5),
		RateLimitToken:   getEnvAsInt("RATE_LIMIT_TOKEN", 10),
		BlockTime:        time.Duration(getEnvAsInt("BLOCK_TIME", 100)) * time.Second,
		EnableIPLimit:    getEnvAsBool("ENABLE_IP_LIMIT", true),
		EnableTokenLimit: getEnvAsBool("ENABLE_TOKEN_LIMIT", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
