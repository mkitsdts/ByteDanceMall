package redis

import (
	"bytedancemall/auth/config"

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

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.Conf.Redis.Host[0],  // use default Addr
		Password: config.Conf.Redis.Password, // no password set
	})
}
