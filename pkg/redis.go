package pkg

import (
	"bytedancemall/cart/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	// Cluster *redis.ClusterClient
	client *redis.Client
)

func NewRedis() error {
	if len(config.Cfg.Redis.Host) < 1 {
		return fmt.Errorf("redis configuration is incomplete")
	} else if len(config.Cfg.Redis.Host) == 1 {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Cfg.Redis.Host[0], config.Cfg.Redis.Port),
			Password: config.Cfg.Redis.Password,
		})
		// 测试连接
		if err := client.Ping(context.Background()).Err(); err != nil {
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
		return nil
	} else {
		return fmt.Errorf("only single Redis instance is supported currently")
	}
}

func GetRedisCli() *redis.Client {
	return client
}
