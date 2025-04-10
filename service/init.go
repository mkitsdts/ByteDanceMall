package service

import (
	"bytedancemall/seckill/config"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewSeckillService() *SeckillSer {
	// 解析配置文件
	SeckillService := &SeckillSer{}
	file, err := os.Open("configs.json")
	if err != nil {
		panic("Failed to open config file: " + err.Error())
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &config.Configs{}
	err = decoder.Decode(&config)
	if err != nil {
		panic("Failed to decode config file: " + err.Error())
	}

	// 连接 Redis
	// SeckillService.RedisCluster = redis.NewClusterClient(&redis.ClusterOptions{
	// 	Addrs: config.RedisConfig.Host,
	// 	Password: config.RedisConfig.Pass,
	// 	PoolSize: 10,
	// })
	SeckillService.RedisCluster = redis.NewClient(&redis.Options{
		Addr:     config.RedisConfig.Host[0],
		Password: config.RedisConfig.Pass,
		DB:       0, // use default DB
		PoolSize: 10,
	})
	// 测试 Redis 连接
	_, err = SeckillService.RedisCluster.Ping(context.Background()).Result()
	if err != nil {
		panic("Failed to ping Redis: " + err.Error())
	}
	fmt.Println("Redis connected successfully")
	return SeckillService
}