package pkg

import (
	"bytedancemall/inventory/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClusterClient() (*redis.ClusterClient, error) {

	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    config.Cfg.Redis.Host,
		Password: config.Cfg.Redis.Password,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}
	return redisClient, nil
}

func NewRedisClient() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.Cfg.Redis.Host[0], config.Cfg.Redis.Port),
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return redisClient, nil
}
