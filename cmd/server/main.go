package main

import (
	"log"
	"net/http"
	"rate-limiter/internal/handler"
	"rate-limiter/internal/repository"
)

func main() {
	mux := http.NewServeMux()
	redisClient := repository.NewRedisClient()
	handler.SetRedis(redisClient)
	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/test-redis", handler.TestRedisHandler)
	mux.HandleFunc("/rate-limit", handler.RateLimitHandler)

	log.Println("Server running on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
