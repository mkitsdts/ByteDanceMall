package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func (s *SeckillSer) AsyncEnsureRedisClientSet(key string, value any) {
	MAX := 10
	gid := getGid()
	slog.Info("AsyncEnsureRedisSet", "key", key, "value", value, "goroutineId", gid)
	for i := range MAX {
		result := s.RedisClient.Set(context.Background(), key, value, 0).Err()
		if result == nil {
			slog.Info("Redis set success", "key", key, "value", value, "goroutineId", gid)
			return
		}
		fmt.Println("Redis set failed, retrying...")
		// 指数退避算法
		duration := time.Duration(1<<i) * time.Millisecond
		time.Sleep(duration)
	}
	slog.Error("Redis set failed after retries", "key", key, "value", value, "goroutineId", gid)
}

func (s *SeckillSer) AsyncEnsureRedisClusterSet(key string, value any) {
	MAX := 10
	gid := getGid()
	slog.Info("AsyncEnsureRedisSet", "key", key, "value", value, "goroutineId", gid)
	for i := range MAX {
		err := s.RedisClient.ForEachMaster(context.Background(), func(ctx context.Context, shard *redis.Client) error {
			max := 3
			for range max {
				result := shard.Set(ctx, key, value, 0).Err()
				if result == nil {
					slog.Info("Redis set success", "key", key, "value", value, "goroutineId", gid)
					break
				}
			}
			slog.Error("Redis set failed after retries", "key", key, "value", value, "goroutineId", gid)
			return nil
		})
		if err == nil {
			slog.Info("Redis set success", "key", key, "value", value, "goroutineId", gid)
			return
		}
		fmt.Println("Redis set failed, retrying...")
		// 指数退避算法
		duration := time.Duration(1<<i) * time.Millisecond
		time.Sleep(duration)
	}
	slog.Error("Redis set failed after retries", "key", key, "value", value, "goroutineId", gid)
}

func (s *SeckillSer) AsyncEnsureRedisZAdd(key string, value redis.Z) {
	MAX := 10
	gid := getGid()
	slog.Info("AsyncEnsureRedisZAdd", "key", key, "value", value, "goroutineId", gid)
	for i := range MAX {
		result := s.RedisClient.ZAdd(context.Background(), key, value).Err()
		if result == nil {
			fmt.Println("Redis set success", "goroutineId", gid)
			return
		}
		fmt.Println("Redis set failed, retrying...")
		// 指数退避算法
		duration := time.Duration(1<<i) * time.Millisecond
		time.Sleep(duration)
	}
	slog.Error("Redis set failed after retries", "key", key, "value", value, "goroutineId", gid)
}

func (s *SeckillSer) AsyncEnsureKafkaAdd(ctx context.Context, ProductId uint32, UserId uint32) error {
	// 将请求放入MQ中
	message := kafka.Message{
		Key:   fmt.Appendf(nil, "seckill_queue_%d", ProductId),
		Value: fmt.Appendf(nil, "%d", UserId),
	}
	return s.KafkaProducer.WriteMessages(ctx, message)
}
