package redis

import (
	"bytedancemall/seckill/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClusterClient(cfg *config.Config) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.Redis.Host,
		MaxIdleConns: 10,
	})
	for range 30 {
		if _, err := client.Ping(context.Background()).Result(); err == nil {
			return client, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	slog.Error("Failed to ping Redis")
	return nil, fmt.Errorf("failed to connect to Redis cluster")
}
