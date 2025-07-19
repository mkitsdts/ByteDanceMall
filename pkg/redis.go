package pkg

import (
	"bytedancemall/user/config"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(configs *config.Config) (*redis.ClusterClient, error) {
	// 生成redis集群的地址
	var redisAddrs []string
	for _, v := range configs.RedisConfig.Configs {
		redisAddrs = append(redisAddrs, v.Host+":"+v.Port)
	}
	// 初始化redi
	s := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        redisAddrs,
		Password:     configs.RedisConfig.Password,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	// 测试redis连接
	if _, err := s.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}
	return s, nil
}
