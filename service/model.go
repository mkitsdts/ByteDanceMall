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
	slog.Debug("Kafka ", "host", config.KafkaConfig.Brokers[0], "topic", config.KafkaConfig.Topic)
	// 初始化 Kafka 生产者
	SeckillService.KafkaProducer = &kafka.Writer{
		Addr:                   kafka.TCP(config.KafkaConfig.Brokers[0]),
		Topic:                  config.KafkaConfig.Topic,
		AllowAutoTopicCreation: true,
	}
	fmt.Println("Init success")
	return SeckillService
}
