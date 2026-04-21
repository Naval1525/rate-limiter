package main

import (
	"log"
	"net/http"

	"rate-limiter/internal/handler"
	"rate-limiter/internal/middleware"
	"rate-limiter/internal/repository"
	"rate-limiter/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	mux := http.NewServeMux()

	// Redis
	redisClient := repository.NewRedisSimpleClient()
	// Lua
	lua := repository.NewLuaScript()

	// Kafka
	go service.StartConsumer()

	// Middleware
	rateLimiter := middleware.NewRateLimiter(lua, redisClient)
	mux.Handle("/metrics", promhttp.Handler())

	// Routes
	mux.HandleFunc("/health", handler.HealthHandler)

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello 🚀"))
	})

	// Wrap mux with middleware
	handlerWithMiddleware := rateLimiter.Middleware(mux)

	log.Println("Server running on :8080")

	err := http.ListenAndServe(":8080", handlerWithMiddleware)
	if err != nil {
		log.Fatal(err)
	}
}
