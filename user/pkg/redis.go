package pkg

import (
	"bytedancemall/user/config"
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.ClusterClient, error) {
	// 初始化redis
	s := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Cfg.RedisConfig.Host,
		Password:     config.Cfg.RedisConfig.Password,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	// 测试redis连接
	if _, err := s.Ping(context.Background()).Result(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}
	slog.Info("Connected to Redis successfully")
	return s, nil
}
