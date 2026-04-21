package repository

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

func NewKafkaProducer() *KafkaProducer {
	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP("kafka:9092"),
			Topic:    "rate-limit-events",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (kp *KafkaProducer) SendEvent(user string) error {
	return kp.Writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(user),
			Value: []byte(time.Now().String()),
		},
	)
}
