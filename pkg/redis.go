package pkg

import (
	"bytedancemall/inventory/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClusterClient(cfg *config.RedisConfig) (*redis.ClusterClient, error) {
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    cfg.Host,
		Password: cfg.Password,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}
	return redisClient, nil
}
