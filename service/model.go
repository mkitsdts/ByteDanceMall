package service

import (
	"bytedancemall/seckill/config"
	seckill "bytedancemall/seckill/proto"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

const (
	layout string = "2006-01-02 15:04:05"
)

type SeckillSer struct {
	RedisClient *redis.ClusterClient
	// RedisClient   *redis.Client // 这里使用单机版 Redis 客户端
	KafkaProducer *kafka.Writer
	KafkaReader   *kafka.Reader
	seckill.UnimplementedSeckillServiceServer
}

func NewSeckillService() *SeckillSer {
	// 解析配置文件
	SeckillService := &SeckillSer{}
	file, err := os.Open("configs.json")
	if err != nil {
		slog.Error("Failed to open config file: " + err.Error())
		return nil
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &config.Configs{}
	err = decoder.Decode(&config)
	if err != nil {
		slog.Error("Failed to decode config file: " + err.Error())
		return nil
	}

	if len(config.RedisConfig.Host) == 0 || len(config.Producer.Brokers) == 0 || len(config.Reader.Brokers) == 0 {
		slog.Error("config is not configured")
		return nil
	}
	// 初始化 Redis 集群
	SeckillService.RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.RedisConfig.Host,
		MaxIdleConns: 10,
	})
	// 测试 Redis 连接
	_, err = SeckillService.RedisClient.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("Failed to ping Redis: " + err.Error())
		return nil
	}

	/*
		SeckillService.RedisClient = redis.NewClient(&redis.Options{
			Addr:     config.RedisConfig.Host[0],
			Password: config.RedisConfig.Pass,
			DB:       0, // use default DB
			PoolSize: 10,
		})
		// 测试 Redis 连接
		_, err = SeckillService.RedisClient.Ping(context.Background()).Result()
		if err != nil {
			panic("Failed to ping Redis: " + err.Error())
		}
	*/

	slog.Debug("Kafka ", "host", config.Producer.Brokers[0], "topic", config.Producer.Topic)
	// 初始化 Kafka 生产者
	SeckillService.KafkaProducer = &kafka.Writer{
		Addr:                   kafka.TCP(config.Producer.Brokers[0]),
		Topic:                  config.Producer.Topic,
		AllowAutoTopicCreation: true,
	}
	// 初始化 Kafka 消费者
	SeckillService.KafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Reader.Brokers,
		Topic:    config.Reader.Topic,
		GroupID:  config.Reader.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	fmt.Println("Init success")
	return SeckillService
}
