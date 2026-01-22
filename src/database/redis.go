package database

import (
	"ambassador/src/config"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var Redis *redis.Client

// ConnectRedis creates Redis connection (called from db.go)
func ConnectRedis(cfg *config.Config) error {
    Redis = redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
        Password:     cfg.RedisPassword,
        DB:           0,
        PoolSize:     10,
        MinIdleConns: 2,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout:  3 * time.Second,
    })

    ctx := context.Background()
    pong, err := Redis.Ping(ctx).Result()
    if err != nil {
        return fmt.Errorf("Redis connection failed: %w", err)
    }
    fmt.Println("Redis connected:", pong)
    return nil
}

// Cache methods
func CacheSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return Redis.Set(ctx, key, value, ttl).Err()
}

func CacheGet(ctx context.Context, key string) (string, error) {
    return Redis.Get(ctx, key).Result()
}

func CacheDelete(ctx context.Context, key string) error {
    return Redis.Del(ctx, key).Err()
}