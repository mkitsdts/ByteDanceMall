package service

import (
	"bytedancemall/seckill/config"
	seckill "bytedancemall/seckill/proto"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

const (
	layout string = "2006-01-02 15:04:05"
)

type SeckillSer struct {
	//RedisCluster *redis.ClusterClient
	RedisCluster  *redis.Client // 这里使用单机版 Redis 客户端
	KafkaProducer *kafka.Writer
	seckill.UnimplementedSeckillServiceServer
}

func NewSeckillService() *SeckillSer {
	// 解析配置文件
	SeckillService := &SeckillSer{}
	file, err := os.Open("configs.json")
	if err != nil {
		panic("Failed to open config file: " + err.Error())
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &config.Configs{}
	err = decoder.Decode(&config)
	if err != nil {
		panic("Failed to decode config file: " + err.Error())
	}

	SeckillService.RedisCluster = redis.NewClient(&redis.Options{
		Addr:     config.RedisConfig.Host[0],
		Password: config.RedisConfig.Pass,
		DB:       0, // use default DB
		PoolSize: 10,
	})
	// 测试 Redis 连接
	_, err = SeckillService.RedisCluster.Ping(context.Background()).Result()
	if err != nil {
		panic("Failed to ping Redis: " + err.Error())
	}
	fmt.Println("Redis connected successfully")

	SeckillService.KafkaProducer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      config.KafkaConfig.Brokers,
		Topic:        config.KafkaConfig.Topic,
		Balancer:     &kafka.Hash{},
		WriteTimeout: 1 * time.Second,
	})
	// 测试 Kafka 连接
	err = SeckillService.KafkaProducer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("test"),
		Value: []byte("test"),
	})
	if err != nil {
		panic("Failed to write message to Kafka: " + err.Error())
	}
	fmt.Println("Kafka connected successfully")
	return SeckillService
}
