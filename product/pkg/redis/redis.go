package redis

import (
	"bytedancemall/product/config"
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var cluster *redis.ClusterClient

func Get() *redis.ClusterClient {
	return cluster
}

func NewRedisCluster() error {
	cluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: config.GetConfig().Redis.Host,
	})
	for range 30 {
		if err := cluster.Ping(context.Background()).Err(); err != nil {
			slog.Error("failed to connect to redis cluster", "error", err)
			return err
		}
		time.Sleep(100 * time.Millisecond) // wait for the cluster to stabilize
	}
	return nil
}
