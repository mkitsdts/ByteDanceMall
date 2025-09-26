package service

import (
	"fmt"
	"log/slog"
	"time"

	"bytedancemall/product/pkg/redis"

	"golang.org/x/net/context"
)

func (s *ProductService) asyncEnsureRedisSet(ctx context.Context, key string, value any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		maxRetries := 3
		for i := range maxRetries {
			if i > 0 {
				// 如果不是第一次尝试，等待一段时间再重试
				time.Sleep(time.Duration(i) * time.Second)
				slog.Info("Retrying to set value in Redis", "key", key, "attempt", i+1)
			}

			err := redis.Get().Set(ctx, key, value, 0).Err()
			if err == nil {
				slog.Info("Successfully set value in Redis", "key", key, "value", value)
				return nil // 成功设置值
			}

			// 如果是最后一次尝试，返回错误
			if i == maxRetries-1 {
				slog.Error("Failed to set value in Redis after retries", "key", key, "error", err)
				return fmt.Errorf("failed to set value in Redis after %d attempts: %w", maxRetries, err)
			}
		}
	}

	return nil
}
