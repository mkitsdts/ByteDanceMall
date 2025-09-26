package redis

import (
	"bytedancemall/order/config"
	"context"
	"fmt"
	"time"

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
			Addr:     config.Cfg.Redis.Host[0],
			Password: config.Cfg.Redis.Password,
			// 参数
			MaxRetries:   3,
			MinIdleConns: 10,
			PoolSize:     50,
			PoolTimeout:  2 * time.Second,
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
