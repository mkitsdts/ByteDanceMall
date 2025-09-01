package pkg

import (
	"bytedancemall/order/config"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	readers = make(map[string]*kafka.Reader)
	writers = make(map[string]*kafka.Writer)
)

func GetReader(tag string) *kafka.Reader {
	if reader, ok := readers[tag]; ok {
		return reader
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Cfg.KafkaReader.Host,
		Topic:    tag,
		MaxWait:  10 * time.Second,
		MaxBytes: 10e6,
		GroupID:  config.Cfg.KafkaReader.GroupID,
	})
	readers[tag] = reader
	return reader
}

func GetWriter(tag string) *kafka.Writer {
	if writer, ok := writers[tag]; ok {
		return writer
	}
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: config.Cfg.KafkaWriter.Host,
		Topic:   tag,
	})
	writers[tag] = writer
	return writer
}

func NewKafkaWriter() error {
	if err := ensureTopics(config.Cfg.KafkaReader.Host, config.Cfg.KafkaReader.Topic); err != nil {
		return fmt.Errorf("failed to ensure kafka topics: %w", err)
	}

	writers := make(map[string]*kafka.Writer)

	for _, topic := range config.Cfg.KafkaReader.Topic {
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers: config.Cfg.KafkaReader.Host,
			Topic:   topic,
		})
		writers[topic] = writer
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
	readers := make(map[string]*kafka.Reader)
	for _, topic := range config.Cfg.KafkaReader.Topic {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  config.Cfg.KafkaReader.Host,
			Topic:    topic,
			MaxWait:  10 * time.Second,
			MaxBytes: 10e6,
			GroupID:  config.Cfg.KafkaReader.GroupID,
		})
		readers[topic] = reader
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
