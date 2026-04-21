package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"rate-limiter/internal/repository"
)

type RateLimiter struct {
	Lua   *repository.LuaScript
	Redis *repository.RedisClient
}

func NewRateLimiter(lua *repository.LuaScript, redis *repository.RedisClient) *RateLimiter {
	return &RateLimiter{
		Lua:   lua,
		Redis: redis,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		apiKey := r.Header.Get("X-API-Key")
		capacity := 5.0
		rate := 2.0
		if apiKey == "premium" {
			capacity = 20
			rate = 10
		}
		if apiKey == "" {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			apiKey = ip
		}

		requestsAllowed.Inc()
		allowed, err := rl.Lua.Allow(ctx, rl.Redis.Client, apiKey, capacity, rate)
		if err != nil {
			http.Error(w, "Internal error", 500)
			return
		}

		if !allowed {
			requestsBlocked.Inc()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
