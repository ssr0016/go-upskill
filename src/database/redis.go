package database

import (
	"ambassador/src/config"
	"context"
	"fmt"
	"log"
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

// ClearProductCaches invalidates ALL product-related caches
// Call this after Create, Update, or Delete operations
func ClearProductCaches(ctx context.Context, productIDs ...uint) {
    // Run in background goroutine to not block API response
    go func() {
        startTime := time.Now()
        deletedCount := 0
        
        log.Printf("Starting cache invalidation for product IDs: %v", productIDs)
        
        // ================================================================
        // STEP 1: Delete specific known cache keys
        // ================================================================
        specificKeys := []string{
            "products_frontend:all",  // Frontend listing
            "products_admin",         // Admin listing
            "products:all:v2",        // Backend all products
        }
        
        // Add individual product caches
        for _, id := range productIDs {
            specificKeys = append(specificKeys, fmt.Sprintf("product:%d", id))
        }
        
        // Delete each specific key
        for _, key := range specificKeys {
            err := Redis.Del(ctx, key).Err()
            if err != nil && err != redis.Nil {
                log.Printf("Failed to delete %s: %v", key, err)
            } else {
                deletedCount++
                log.Printf("Deleted: %s", key)
            }
        }
        
        // ================================================================
        // STEP 2: Delete ALL paginated caches using pattern matching
        // ================================================================
        patterns := []string{
            "products:p:*",      // All paginated caches (legacy)
            "products:v2:p:*",   // All v2 paginated caches
        }
        
        for _, pattern := range patterns {
            // Find all keys matching the pattern
            keys, err := Redis.Keys(ctx, pattern).Result()
            if err != nil {
                log.Printf("Failed to find keys for pattern %s: %v", pattern, err)
                continue
            }
            
            // Delete all found keys in batch
            if len(keys) > 0 {
                deleted, err := Redis.Del(ctx, keys...).Result()
                if err != nil {
                    log.Printf("Failed to delete keys for pattern %s: %v", pattern, err)
                } else {
                    deletedCount += int(deleted)
                    log.Printf("Deleted %d keys matching pattern: %s", deleted, pattern)
                }
            } else {
                log.Printf("No keys found for pattern: %s", pattern)
            }
        }
        
        duration := time.Since(startTime)
        log.Printf(" Cache invalidation complete! Deleted %d keys in %v (Product IDs: %v)", 
            deletedCount, duration, productIDs)
    }()
}

// Optional: Add a synchronous version if you need to wait for completion
func ClearProductCachesSync(ctx context.Context, productIDs ...uint) error {
    deletedCount := 0
    
    // Delete specific keys
    specificKeys := []string{
        "products_frontend:all",
        "products_admin",
        "products:all:v2",
    }
    
    for _, id := range productIDs {
        specificKeys = append(specificKeys, fmt.Sprintf("product:%d", id))
    }
    
    for _, key := range specificKeys {
        if err := Redis.Del(ctx, key).Err(); err == nil {
            deletedCount++
        }
    }
    
    // Delete paginated caches
    patterns := []string{"products:p:*", "products:v2:p:*"}
    for _, pattern := range patterns {
        keys, err := Redis.Keys(ctx, pattern).Result()
        if err != nil {
            continue
        }
        if len(keys) > 0 {
            deleted, _ := Redis.Del(ctx, keys...).Result()
            deletedCount += int(deleted)
        }
    }
    
    log.Printf("Cleared %d product caches synchronously", deletedCount)
    return nil
}