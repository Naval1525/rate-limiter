package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsAllowed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requests_allowed_total",
			Help: "Total allowed requests",
		},
	)

	requestsBlocked = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requests_blocked_total",
			Help: "Total blocked requests",
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(requestsAllowed, requestsBlocked)
}
