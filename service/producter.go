package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *SeckillSer) AddItemHandler(ctx context.Context, ProductId uint32, Quantity uint32, ReleaseTime string) {
	t, _ := time.Parse(ReleaseTime, layout)
	// 秒杀商品基本信息
	productInfoKey := fmt.Sprintf("seckill:product:%d", ProductId)
	productInfo, _ := json.Marshal(map[string]any{
		"id":          ProductId,
		"quantity":    Quantity,
		"releaseTime": t.Unix(),
	})
	fmt.Println("productInfo", productInfo)

	go s.AsyncEnsureRedisClusterSet(productInfoKey, productInfo)

	// 设置总库存到Redis
	Quantity = Quantity / REDIS_CLUSTER_COUNT
	stockKey := fmt.Sprintf("seckill:stock:%d", ProductId)
	go s.AsyncEnsureRedisClusterSet(stockKey, Quantity)

	// 设置秒杀开始时间标记
	releaseTimeKey := fmt.Sprintf("seckill:releaseTime:%d", ProductId)
	fmt.Println("设置秒杀开始时间标记", releaseTimeKey)
	go s.AsyncEnsureRedisClusterSet(releaseTimeKey, t.Unix())

	// 添加到秒杀商品列表
	go s.AsyncEnsureRedisZAdd("seckill:products", redis.Z{
		Score:  float64(t.Unix()),
		Member: ProductId,
	})
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
	stockKey := fmt.Sprintf("seckill:stock:%d", ProductId)
	orderKey := fmt.Sprintf("seckill:orders:%d", ProductId)

	fmt.Println("秒杀商品", ProductId, "用户", UserId)
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

	fmt.Println("秒杀结果", result)
	if result.(int64) == 0 {
		fmt.Println("秒杀失败")
		return nil
	}

	fmt.Println("放入队列")
	err = s.AsyncEnsureKafkaAdd(ctx, ProductId, UserId)
	if err != nil {
		// 记录错误
		slog.Error("添加到队列失败", "error", err, "productId", ProductId, "userId", UserId)

		// 恢复Redis状态
		s.RedisClient.Incr(ctx, stockKey)
		s.RedisClient.SRem(ctx, orderKey, UserId)
		return fmt.Errorf("添加到队列失败: %w", err)
	}

	return nil
}
