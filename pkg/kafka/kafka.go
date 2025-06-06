package kafka

import (
	"bytedancemall/seckill/internal/config"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

func InitKafkaProducer(cfg *config.Config) (*kafka.Writer, error) {
	if len(cfg.Kafka.Brokers) == 0 || len(cfg.Kafka.Topic) == 0 {
		return nil, fmt.Errorf("kafka configuration is incomplete")
	}

	producer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers[0]),
		Topic:                  cfg.Kafka.Topic,
		AllowAutoTopicCreation: true,
		Transport: &kafka.Transport{
			// 确保禁用DNS解析缓存或本地地址转换
			DialTimeout: 5 * time.Second,
			TLS:         nil, // 如果不需要TLS
		},
	}

	for range 30 {
		_, err := net.Dial("tcp", cfg.Kafka.Brokers[0])
		if err == nil {
			return producer, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("failed to connect to Kafka")
}
