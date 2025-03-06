package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var CacheClient *redis.Client

// Initialize Redis Client
func InitCache() {
	CacheClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis container
		Password: "",               // No password by default
		DB:       0,                // Default DB
	})

	// Test connection
	_, err := CacheClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("❌ Redis connection failed:", err)
	} else {
		fmt.Println("✅ Connected to Redis")
	}
}

// Set cache with expiration
func SetCache(key string, value string, duration time.Duration) error {
	return CacheClient.Set(ctx, key, value, duration).Err()
}

// Get cache
func GetCache(key string) (string, error) {
	return CacheClient.Get(ctx, key).Result()
}

// Delete cache
func DeleteCache(key string) error {
	return CacheClient.Del(ctx, key).Err()
}
