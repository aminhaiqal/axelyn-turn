package db

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    Client *redis.Client
}

// NewRedisClient reads UPSTASH_REDIS_URL from env (rediss://user:token@host:port)
// and returns a connected *RedisClient.
func NewRedisClient() *RedisClient {
    redisURL := os.Getenv("UPSTASH_REDIS_URL")
    if redisURL == "" {
        log.Fatal("Missing UPSTASH_REDIS_URL in environment")
    }

    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        log.Fatalf("Invalid UPSTASH_REDIS_URL: %v", err)
    }

    rdb := redis.NewClient(opt)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Test connection
    if _, err := rdb.Ping(ctx).Result(); err != nil {
        log.Fatalf("Failed to connect to Upstash Redis: %v", err)
    }

    log.Println("Connected to Upstash Redis")
    return &RedisClient{Client: rdb}
}
