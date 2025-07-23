package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/segmentio/kafka-go"
)

// 下单
func (s *OrderService) CreateOrder() {
	for {
		// 从 Kafka 中读取订单消息
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg, err := s.Reader.ReadMessage(ctx)
		if err != nil {
			cancel()
			slog.Error("Failed to read message from Kafka", "error", err)
			time.Sleep(time.Second) // 避免CPU空转
			continue
		}

		if err = s.ConsumeCreateMsg(ctx, msg); err != nil {
			slog.Error("Failed to consume create message", "error", err)
			cancel()
			continue
		}

		if err := s.Reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("Failed to commit message", "error", err)
		}
		cancel()
	}
}

func (s *OrderService) ConsumeCreateMsg(ctx context.Context, msg kafka.Message) error {
	// 解析订单消息
	var order Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		slog.Error("Failed to unmarshal order message", "error", err)
		return err
	}
	order.State = "WAITING_PAYMENT"
	// 获取分布式锁
	mutex := s.Redsync.NewMutex("order_mutex:"+order.OrderId, redsync.WithExpiry(5*time.Second), redsync.WithTries(3))
	if err := mutex.Lock(); err != nil {
		slog.Error("Failed to acquire lock", "error", err)
		return err
	}
	defer mutex.Unlock()

	// 将订单存入数据库
	if err := s.Db.Master.Create(&order).Error; err != nil {
		slog.Error("Failed to create order in database", "error", err)
		return err
	}
	return nil
}

func (s *OrderService) ConsumeUpdateMsg(ctx context.Context, msg kafka.Message) error {
	// 解析订单状态更新消息
	var orderUpdate Order
	if err := json.Unmarshal(msg.Value, &orderUpdate); err != nil {
		slog.Error("Failed to unmarshal order update message", "error", err)
		return err
	}

	// 获取分布式锁
	mutex := s.Redsync.NewMutex("order_mutex:"+orderUpdate.OrderId, redsync.WithExpiry(5*time.Second), redsync.WithTries(3))
	if err := mutex.Lock(); err != nil {
		slog.Error("Failed to acquire lock", "error", err)
		return err
	}
	defer mutex.Unlock()

	// 更新订单状态
	if err := s.Db.Master.Model(&Order{}).Where("order_id = ?", orderUpdate.OrderId).Update("state", orderUpdate.State).Error; err != nil {
		slog.Error("Failed to update order state in database", "error", err)
		return err
	}
	return nil
}
