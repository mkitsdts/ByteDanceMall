package redis

import (
	"bytedancemall/inventory/config"
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func NewClusterClient() (*goredis.ClusterClient, error) {
	redisClient := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:    config.Cfg.Redis.Host,
		Password: config.Cfg.Redis.Password,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}
	return redisClient, nil
}

func NewClient() (*goredis.Client, error) {
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%s", config.Cfg.Redis.Host[0], config.Cfg.Redis.Port),
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return redisClient, nil
}
