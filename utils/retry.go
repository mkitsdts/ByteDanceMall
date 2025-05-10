package utils

import (
	"context"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// 辅助函数：带重试的Redis操作
func RetryRedisOperation(ctx context.Context, operation func() error) error {
	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil || err == redis.Nil {
			return err
		}

		// 指数退避
		backoff := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			continue
		}
	}

	return err
}
