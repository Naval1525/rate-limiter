package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"rate-limiter/internal/repository"
)

var redisClient *repository.RedisClient

func SetRedis(r *repository.RedisClient) {
	redisClient = r
}

func TestRedisHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "test_key"
	value := "hello_redis"

	// SET
	err := redisClient.Client.Set(ctx, key, value, 0).Err()
	if err != nil {
		http.Error(w, "Redis SET failed", 500)
		return
	}

	// GET
	val, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		http.Error(w, "Redis GET failed", 500)
		return
	}

	fmt.Fprintf(w, "Stored value: %s", val)
}
