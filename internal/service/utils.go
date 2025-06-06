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
	slog.Info("AsyncEnsureRedisSet", "key", key, "value", value)
	for i := range MAX {
		result := s.RedisClient.Set(context.Background(), key, value, 0).Err()
		if result == nil {
			slog.Info("Redis set success", "key", key, "value", value)
			return
		}
		slog.Warn("Redis set failed, retrying...")
		// 指数退避算法
		duration := time.Duration(1<<i) * time.Millisecond
		time.Sleep(duration)
	}
	slog.Error("Redis set failed after retries", "key", key, "value", value)
}

func (s *SeckillSer) AsyncEnsureRedisClusterSet(key string, value any) {
	MAX := 10
	slog.Info("AsyncEnsureRedisSet", "key", key, "value", value)
	for i := range MAX {
		err := s.RedisClient.ForEachMaster(context.Background(), func(ctx context.Context, shard *redis.Client) error {
			max := 3
			for range max {
				result := shard.Set(ctx, key, value, 0).Err()
				if result == nil {
					slog.Info("Redis set success", "key", key, "value", value)
					break
				}
			}
			slog.Error("Redis set failed after retries", "key", key, "value", value)
			return nil
		})
		if err == nil {
			slog.Info("Redis set success", "key", key, "value", value)
			return
		}
		slog.Warn("Redis set failed, retrying...")
		// 指数退避算法
		duration := time.Duration(1<<i) * time.Millisecond
		time.Sleep(duration)
	}
	slog.Error("Redis set failed after retries", "key", key, "value", value)
}

func (s *SeckillSer) AsyncEnsureRedisZAdd(key string, value redis.Z) {
	MAX := 10
	slog.Info("AsyncEnsureRedisZAdd", "key", key, "value", value)
	for i := range MAX {
		result := s.RedisClient.ZAdd(context.Background(), key, value).Err()
		if result == nil {
			slog.Info("Redis set success")
			return
		}
		slog.Warn("Redis set failed, retrying...")
		// 指数退避算法
		duration := time.Duration(1<<i) * time.Millisecond
		time.Sleep(duration)
	}
	slog.Error("Redis set failed after retries", "key", key, "value", value)
}

func (s *SeckillSer) AsyncEnsureKafkaAdd(ctx context.Context, ProductId uint32, UserId uint32) error {
	// 将请求放入MQ中
	message := kafka.Message{
		Key:   fmt.Appendf(nil, "seckill_queue_%d", ProductId),
		Value: fmt.Appendf(nil, "%d", UserId),
	}
	err := s.KafkaProducer.WriteMessages(ctx, message)
	if err != nil {
		slog.Error("Failed to write message to Kafka", "error", err)
		return err
	}
	return nil
}

func (s *SeckillSer) AddItemHandler(ctx context.Context, ProductId uint32, Quantity uint32, ReleaseTime string) {
	t, _ := time.Parse(ReleaseTime, layout)
	// 设置总库存
	stockKey := fmt.Sprintf("{seckill:%d}:stock", ProductId)
	go s.AsyncEnsureRedisClientSet(stockKey, Quantity)

	// 设置秒杀开始时间
	releaseTimeKey := fmt.Sprintf("{seckill:%d}:releaseTime", ProductId)
	slog.Info(releaseTimeKey)
	go s.AsyncEnsureRedisClientSet(releaseTimeKey, t.Unix())
}

// 秒杀商品
func (s *SeckillSer) TrySecKillItemHandler(ctx context.Context, ProductId uint32, UserId uint32) error {
	// 检查秒杀是否已开始
	// releaseTimeKey := fmt.Sprintf("seckill:releasetime:%d", ProductId)

	// maxRetry := 3
	// var releaseTimeStr string
	// var err error
	// for i := range maxRetry {
	// 	releaseTimeStr, err = s.RedisCluster.Get(ctx, releaseTimeKey).Result()
	// 	if err == nil {
	// 		break
	// 	}
	// 	if err == redis.Nil {
	// 		return fmt.Errorf("no such product: %d", ProductId)
	// 	}
	// 	// 如果Redis中没有数据，可能是因为数据还未设置，稍等一会儿再试
	// 	fmt.Println("Redis key not found, retrying...")
	// 	time.Sleep(time.Duration(1<<i) * time.Millisecond)
	// }

	// releaseTime, _ := strconv.ParseInt(releaseTimeStr, 10, 64)
	// now := time.Now().Unix()

	// if now < releaseTime {
	// 	// return fmt.Errorf("秒杀未开始")
	// }

	// 秒杀逻辑 - 使用Lua脚本保证原子性
	stockKey := fmt.Sprintf("{seckill:%d}:stock", ProductId)
	orderKey := fmt.Sprintf("{seckill:%d}:orders", ProductId)

	slog.Info("秒杀商品", "productId", ProductId, "userId", UserId)
	script := `
	local stock = redis.call('get', KEYS[1])
	if not stock or tonumber(stock) <= 0 then return 0 end
	
	redis.call('decr', KEYS[1])
	redis.call('sadd', KEYS[2], ARGV[1])
	return 1
	`

	result, err := s.RedisClient.Eval(ctx, script, []string{stockKey, orderKey}, UserId).Result()
	if err != nil {
		return err
	}

	slog.Info("秒杀结果", "result", result)
	if result.(int64) == 0 {
		slog.Error("秒杀失败")
		return nil
	}

	slog.Info("放入队列")
	go s.AsyncEnsureKafkaAdd(ctx, ProductId, UserId)

	return nil
}
