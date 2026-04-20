package handler

import (
	"fmt"
	"net/http"
	"rate-limiter/internal/service"
)

var bucket = service.NewTokenBucket(5, 2)

func RateLimitHandler(w http.ResponseWriter, r *http.Request) {
	if bucket.Allow() {
		fmt.Fprintf(w, "Request allowed")
	} else {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
	}
}
