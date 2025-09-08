package redis

import "github.com/redis/go-redis/v9"

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
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
