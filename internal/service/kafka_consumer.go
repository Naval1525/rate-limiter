package service

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func StartConsumer() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "rate-limit-events",
		GroupID: "rate-limit-group",
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("error:", err)
			continue
		}

		log.Println("Event received:", string(m.Key))
	}
}
