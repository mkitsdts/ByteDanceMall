package pkg

import (
	"bytedancemall/inventory/config"
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(cfg *config.KafkaWriterConfig) (*kafka.Writer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.Host,
		Topic:   cfg.Topic,
	})

	if err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
	}); err != nil {
		return nil, fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return writer, nil
}

func NewKafkaReader(cfg *config.KafkaReaderConfig) (*kafka.Reader, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Host,
		Topic:    cfg.Topic,
		GroupID:  "order_service_group",
		MaxWait:  10 * time.Second,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	if err := reader.SetOffset(kafka.LastOffset); err != nil {
		return nil, fmt.Errorf("failed to set offset for Kafka reader: %w", err)
	}

	return reader, nil
}
