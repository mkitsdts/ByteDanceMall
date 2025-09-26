package payment

import (
	"bytedancemall/payment/pkg/kafka"
	"bytedancemall/payment/pkg/redis"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	k "github.com/segmentio/kafka-go"
)

type Body struct {
	OrderID uint64 `json:"order_id"`
	UserID  uint64 `json:"user_id"`
}

func HandlePaymentNotify(c *gin.Context) {
	result := CreatePaymentResponseBody{}
	if err := c.BindJSON(&result); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	body, err := translateBody(result.Resource.Algorithm, result.Resource.CipherText)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to translate body"})
		return
	}

	// 先判断订单是否在 redis 里，如果在说明是重复通知
	key := "order_payment_result:" + fmt.Sprint(body.OrderID)
	caa, _ := redis.CheckDuplicate(key)
	if caa {
		c.JSON(400, gin.H{"error": "duplicate order"})
		return
	}

	// 尝试获取分布式锁（用redlock或etcd）
	if cmd := redis.GetRedisCli().SetNX(context.Background(), key, body, 30*time.Second); cmd.Err() != nil {
		if !cmd.Val() {
			c.JSON(500, gin.H{"error": "failed to store payment result"})
			return
		}
	}

	maxRetries := 5
	for i := range maxRetries {
		if err := kafka.GetWriter("order_status").WriteMessages(context.Background(), k.Message{
			Key:   fmt.Append(nil, body.OrderID),
			Value: fmt.Append(nil, body.UserID),
		}); err == nil {
			c.JSON(200, gin.H{"status": "success"})
			return
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			slog.Error("!!!failed to write kafka message", "order_id", body.OrderID)
			c.JSON(500, gin.H{"error": "failed to process payment result"})
			go func() {
				// 发送消息到死信队列
				for i := range maxRetries {
					if err := kafka.GetWriter("order_status_dlq").WriteMessages(context.Background(), k.Message{
						Key:   fmt.Append(nil, body.OrderID),
						Value: fmt.Append(nil, body.UserID),
					}); err == nil {
						break
					}
					time.Sleep(10 << i * time.Millisecond)
				}
				slog.Error("!!!failed to write kafka message to dlq", "order_id", body.OrderID)
			}()
		}
	}
}

func translateBody(algorithm string, ciphertext string) (Body, error) {
	id := strings.Index(ciphertext, "order_id:")
	if algorithm == "AEAD_AES_256_GCM" {
		return Body{
			OrderID: uint64(id),
			UserID:  145,
		}, nil
	}
	return Body{
		OrderID: 114514,
		UserID:  1919810,
	}, nil
}
