package redis

import (
	"bytedancemall/auth/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// var redisClusterClient *redis.ClusterClient

func GetCLI() *redis.Client {
	if redisClient == nil {
		InitRedis()
	}
	return redisClient
}

func InitRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.Conf.Redis.Host[0],
		Password: config.Conf.Redis.Password,
		// 参数
		PoolSize:     50,
		MinIdleConns: 10,
		PoolTimeout:  2 * time.Second, // 等待可用连接的最大时间
	})
	return nil
}
