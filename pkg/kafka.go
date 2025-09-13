package pkg

import (
	"bytedancemall/cart/config"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaWriters map[string]*kafka.Writer
	kafkaReaders map[string]*kafka.Reader
)

func NewKafkaWriter() error {
	if err := ensureTopics(config.Cfg.KafkaReader.Host, config.Cfg.KafkaReader.Topic); err != nil {
		return fmt.Errorf("failed to ensure kafka topics: %w", err)
	}

	kafkaWriters = make(map[string]*kafka.Writer)

	for _, topic := range config.Cfg.KafkaReader.Topic {
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers: config.Cfg.KafkaReader.Host,
			Topic:   topic,
		})
		kafkaWriters[topic] = writer
		if err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte("key"),
			Value: []byte("value"),
		}); err != nil {
			return fmt.Errorf("failed to write message to Kafka: %w", err)
		}
	}
	return nil
}

func NewKafkaReader() error {
	if err := ensureTopics(config.Cfg.KafkaReader.Host, config.Cfg.KafkaReader.Topic); err != nil {
		return fmt.Errorf("failed to ensure kafka topics: %w", err)
	}
	kafkaReaders = make(map[string]*kafka.Reader)
	for _, topic := range config.Cfg.KafkaReader.Topic {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  config.Cfg.KafkaReader.Host,
			Topic:    topic,
			MaxWait:  10 * time.Second,
			MaxBytes: 10e6,
			GroupID:  config.Cfg.KafkaReader.GroupID,
		})
		kafkaReaders[topic] = reader
	}
	return nil
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

func GetKafkaWriter(topic string) *kafka.Writer {
	return kafkaWriters[topic]
}

func GetKafkaReader(topic string) *kafka.Reader {
	return kafkaReaders[topic]
}
