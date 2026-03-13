package kafka

import (
	"bytedancemall/inventory/config"
	"fmt"
	"net"
	"time"

	segmentio "github.com/segmentio/kafka-go"
)

func NewWriter() (map[string]*segmentio.Writer, error) {
	if err := ensureTopics(config.Cfg.KafkaReader.Host, config.Cfg.KafkaReader.Topic); err != nil {
		return nil, fmt.Errorf("failed to ensure kafka topics: %w", err)
	}

	writers := make(map[string]*segmentio.Writer)
	for _, topic := range config.Cfg.KafkaReader.Topic {
		writers[topic] = segmentio.NewWriter(segmentio.WriterConfig{
			Brokers: config.Cfg.KafkaReader.Host,
			Topic:   topic,
		})
	}

	return writers, nil
}

func NewReader() (map[string]*segmentio.Reader, error) {
	if err := ensureTopics(config.Cfg.KafkaReader.Host, config.Cfg.KafkaReader.Topic); err != nil {
		return nil, fmt.Errorf("failed to ensure kafka topics: %w", err)
	}

	readers := make(map[string]*segmentio.Reader)
	for _, topic := range config.Cfg.KafkaReader.Topic {
		readers[topic] = segmentio.NewReader(segmentio.ReaderConfig{
			Brokers:  config.Cfg.KafkaReader.Host,
			Topic:    topic,
			MaxWait:  10 * time.Second,
			MaxBytes: 10e6,
			GroupID:  config.Cfg.KafkaReader.GroupID,
		})
	}
	return readers, nil
}

func ensureTopics(brokers []string, topics []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers provided")
	}

	conn, err := segmentio.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := segmentio.Dial("tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := make([]segmentio.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, segmentio.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	if err := controllerConn.CreateTopics(topicConfigs...); err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}
	return nil
}
