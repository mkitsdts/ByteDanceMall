package redis

import (
	"bytedancemall/user/config"
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient() (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:         config.Conf.Redis.Host[0],
		Password:     config.Conf.Redis.Password,
		PoolSize:     50,
		MinIdleConns: 10,
		PoolTimeout:  2 * time.Second,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}
	return client, nil
}
