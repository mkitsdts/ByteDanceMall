package pkg

import (
	"bytedancemall/inventory/config"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(cfg *config.KafkaWriter) (map[string]*kafka.Writer, error) {
	if err := ensureTopics(cfg.Host, cfg.Topic); err != nil {
		return nil, fmt.Errorf("failed to ensure kafka topics: %w", err)
	}

	writers := make(map[string]*kafka.Writer)

	for _, topic := range cfg.Topic {
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers: cfg.Host,
			Topic:   topic,
		})
		writers[topic] = writer
		if err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte("key"),
			Value: []byte("value"),
		}); err != nil {
			return nil, fmt.Errorf("failed to write message to Kafka: %w", err)
		}
	}

	return writers, nil
}

func NewKafkaReader(cfg *config.KafkaReader) (map[string]*kafka.Reader, error) {
	if err := ensureTopics(cfg.Host, cfg.Topic); err != nil {
		return nil, fmt.Errorf("failed to ensure kafka topics: %w", err)
	}
	readers := make(map[string]*kafka.Reader)
	for _, topic := range cfg.Topic {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  cfg.Host,
			Topic:    topic,
			MaxWait:  10 * time.Second,
			MaxBytes: 10e6,
			GroupID:  cfg.GroupID,
		})
		readers[topic] = reader
	}
	return readers, nil
}

// ensureTopics 检查并创建 Kafka 主题
func ensureTopics(brokers []string, topics []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers provided")
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{}
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	return nil
}
