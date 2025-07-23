package pkg

import (
	"bytedancemall/order/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    cfg.RedisConfig.Host,
		Password: cfg.RedisConfig.Password,
	})
	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return client, nil
}
