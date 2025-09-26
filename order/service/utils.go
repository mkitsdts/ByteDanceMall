package service

import (
	"bytedancemall/order/pkg/redis"
	"context"
	"fmt"
	"time"
)

func waitRedisValue(ctx context.Context, key string) (string, error) {

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	cli := redis.GetRedisCli()
	// 先尝试一次
	if v, err := cli.Get(ctx, key).Result(); err == nil && v != "" {
		return v, nil
	}

	timeout, ok := ctx.Deadline()
	if !ok {
		timeout = time.Now().Add(time.Second * 10) // 默认10秒超时
	}
	for {
		select {
		case <-time.After(time.Until(timeout)):
			return "", fmt.Errorf("wait redis key %s timeout", key)
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			v, err := cli.Get(ctx, key).Result()
			if err == nil && v != "" {
				return v, nil
			}
			// rds.Nil 表示 key 不存在，继续等
		}
	}
}
