package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	})
}

func Cache[T any](key string, value T, seconds int) error {
	return rdb.Set(ctx, fmt.Sprintf("cache:%s", key), value, time.Duration(seconds)*time.Second).Err()
}

func GetCache[T any](key string, target T) error {
	r := rdb.Get(ctx, fmt.Sprintf("cache:%s", key))

	err := r.Scan(target)

	if err == redis.Nil {
		return nil
	}

	return err
}

func Store[T any](key string, value T) error {
	return rdb.Set(ctx, fmt.Sprintf("store:%s", key), value, 0).Err()
}

func GetStore[T any](key string, target T) error {
	r := rdb.Get(ctx, fmt.Sprintf("store:%s", key))

	err := r.Scan(target)

	if err == redis.Nil {
		return nil
	}

	return err
}

func getEnvOrDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
