package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"rate-limiter/internal/repository"
)

var luaScript *repository.LuaScript

func SetLuaScript(l *repository.LuaScript) {
	luaScript = l
}

func clientKey(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return "rate_limit:" + strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return "rate_limit:" + host
}

func RateLimitHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	allowed, err := luaScript.Allow(ctx, redisClient.Client, clientKey(r), 5, 2)
	if err != nil {
		http.Error(w, "Redis error", 500)
		return
	}

	if allowed {
		fmt.Fprintf(w, "✅ Allowed")
	} else {
		http.Error(w, "❌ Rate limit exceeded", 429)
	}
}
