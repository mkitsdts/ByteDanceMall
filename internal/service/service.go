package service

import (
	seckill "bytedancemall/seckill/proto"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type SeckillSer struct {
	RedisClient   *redis.ClusterClient
	KafkaProducer *kafka.Writer
	seckill.UnimplementedSeckillServiceServer
}

func NewSeckillService(redis *redis.ClusterClient, writer *kafka.Writer) (*SeckillSer, error) {
	if redis == nil || writer == nil {
		return nil, fmt.Errorf("redis client, kafka writer or reader cannot be nil")
	}

	service := &SeckillSer{
		RedisClient:   redis,
		KafkaProducer: writer,
	}
	return service, nil
}
