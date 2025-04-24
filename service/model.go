package service

import (
	seckill "bytedancemall/seckill/proto"

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
