package service

import (
	seckill "bytedancemall/seckill/proto"

	"github.com/redis/go-redis/v9"
)

type SeckillSer struct {
	//RedisCluster *redis.ClusterClient
	RedisCluster *redis.Client
	// 这里使用单机版 Redis 客户端
	seckill.UnimplementedSeckillServiceServer
}