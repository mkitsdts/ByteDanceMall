package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// AddItem 添加商品到秒杀列表并设置开始时间和库存
// func (s *SeckillSer) AddItem(ctx context.Context, req *seckill.AddItemReq) (*seckill.AddItemResp, error) {
// 	t, _ := time.Parse(req.ReleaseTime, layout)
// 	// 秒杀商品基本信息
// 	productInfoKey := fmt.Sprintf("seckill:product:%d", req.ProductId)
// 	productInfo, _ := json.Marshal(map[string]any{
// 		"id":          req.ProductId,
// 		"quantity":    req.Quantity,
// 		"releaseTime": t.Unix(),
// 	})
// 	s.RedisCluster.Set(ctx, productInfoKey, productInfo, 0)

// 	// 设置总库存到Redis
// 	stockKey := fmt.Sprintf("seckill:stock:%d", req.ProductId)
// 	s.RedisCluster.Set(ctx, stockKey, req.Quantity, 0)

// 	// 分散库存（提高并发性能）
// 	// 将库存分散到多个bucket中，减轻单个key的压力
// 	// bucketCount := 10
// 	// avg := req.Quantity / uint32(bucketCount)
// 	// remainder := req.Quantity % uint32(bucketCount)

// 	// pipe := s.RedisCluster.Pipeline()
// 	// for i := 0; i < bucketCount; i++ {
// 	// 	bucketStock := avg
// 	// 	if i == 0 {
// 	// 		bucketStock += remainder // 将余数放在第一个桶
// 	// 	}
// 	// 	bucketKey := fmt.Sprintf("seckill:stock:%d:bucket:%d", req.ProductId, i)
// 	// 	pipe.Set(ctx, bucketKey, bucketStock, 0)
// 	// }
// 	// _, err = pipe.Exec(ctx)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("设置分散库存失败: %w", err)
// 	// }

// 	// 设置秒杀开始时间标记
// 	releaseTimeKey := fmt.Sprintf("seckill:ReleaseTime:%d", req.ProductId)
// 	s.RedisCluster.Set(ctx, releaseTimeKey, t.Unix(), 0)

// 	// 添加到秒杀商品列表
// 	s.RedisCluster.ZAdd(ctx, "seckill:products", redis.Z{
// 		Score:  float64(t.Unix()),
// 		Member: req.ProductId,
// 	})

// 	return &seckill.AddItemResp{
// 		Result: true,
// 	}, nil
// }

// // 秒杀商品
// func (s *SeckillSer) TrySecKillItem(ctx context.Context, req *seckill.TrySecKillItemReq) (*seckill.TrySecKillItemResp, error) {
// 	// 检查秒杀是否已开始
// 	releaseTimeKey := fmt.Sprintf("seckill:releasetime:%d", req.ProductId)
// 	releaseTimeStr, err := s.RedisCluster.Get(ctx, releaseTimeKey).Result()
// 	if err != nil {
// 		return nil, fmt.Errorf("商品不在秒杀列表")
// 	}

// 	releaseTime, _ := strconv.ParseInt(releaseTimeStr, 10, 64)
// 	now := time.Now().Unix()

// 	if now < releaseTime {
// 		return &seckill.TrySecKillItemResp{
// 			Result: -1,
// 		}, fmt.Errorf("秒杀未开始")
// 	}

// 	// 秒杀逻辑 - 使用Lua脚本保证原子性
// 	stockKey := fmt.Sprintf("seckill:stock:%d", req.ProductId)
// 	orderKey := fmt.Sprintf("seckill:orders:%d", req.ProductId)

// 	script := `
//     local stock = redis.call('get', KEYS[1])
//     if tonumber(stock) <= 0 then return 0 end

//     redis.call('decr', KEYS[1])
//     redis.call('sadd', KEYS[2], ARGV[1])
//     return 1
//     `

// 	result, err := s.RedisCluster.Eval(ctx, script, []string{stockKey, orderKey}, req.UserId).Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if result.(int64) == 0 {
// 		return &seckill.TrySecKillItemResp{
// 			Result: -1,
// 		}, nil
// 	}

// 	// 将请求放入Stream中
// 	queueKey := fmt.Sprintf("seckill:queue:%d", req.ProductId)
// 	_, err = s.RedisCluster.XAdd(ctx, &redis.XAddArgs{
// 		Stream: queueKey,
// 		Values: map[string]any{
// 			"userId":    req.UserId,
// 			"productId": req.ProductId,
// 		},
// 	}).Result()
// 	if err != nil {
// 		return nil, fmt.Errorf("添加到队列失败: %w", err)
// 	}

// 	return &seckill.TrySecKillItemResp{
// 		Result: 0,
// 	}, nil
// }

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
	fmt.Println(productInfo)
	s.RedisCluster.Set(ctx, productInfoKey, productInfo, 0)

	// 设置总库存到Redis
	stockKey := fmt.Sprintf("seckill:stock:%d", ProductId)
	s.RedisCluster.Set(ctx, stockKey, Quantity, 0)

	// 设置秒杀开始时间标记
	releaseTimeKey := fmt.Sprintf("seckill:ReleaseTime:%d", ProductId)
	fmt.Println("设置秒杀开始时间标记", releaseTimeKey)
	s.RedisCluster.Set(ctx, releaseTimeKey, t.Unix(), 0)

	// 添加到秒杀商品列表
	s.RedisCluster.ZAdd(ctx, "seckill:products", redis.Z{
		Score:  float64(t.Unix()),
		Member: ProductId,
	})
}

// 秒杀商品
func (s *SeckillSer) TrySecKillItemHandler(ctx context.Context, ProductId uint32, UserId uint32) error {
	// 检查秒杀是否已开始
	releaseTimeKey := fmt.Sprintf("seckill:releasetime:%d", ProductId)
	releaseTimeStr, err := s.RedisCluster.Get(ctx, releaseTimeKey).Result()
	if err != nil {
		fmt.Println(releaseTimeKey)
		return fmt.Errorf("商品不在秒杀列表")
	}

	releaseTime, _ := strconv.ParseInt(releaseTimeStr, 10, 64)
	now := time.Now().Unix()

	if now < releaseTime {
		return fmt.Errorf("秒杀未开始")
	}

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

	result, err := s.RedisCluster.Eval(ctx, script, []string{stockKey, orderKey}, UserId).Result()
	if err != nil {
		return err
	}

	fmt.Println("秒杀结果", result)
	if result.(int64) == 0 {
		fmt.Println("秒杀失败")
		return nil
	}

	fmt.Println("放入队列")
	// 将请求放入Stream中
	queueKey := fmt.Sprintf("seckill:queue:%d", ProductId)
	err = s.KafkaProducer.WriteMessages(ctx, kafka.Message{
		Key:   fmt.Appendf(nil, "%s", queueKey),
		Value: fmt.Appendf(nil, "%d", UserId),
	})

	if err != nil {
		return fmt.Errorf("添加到队列失败: %w", err)
	}
	return nil
}
