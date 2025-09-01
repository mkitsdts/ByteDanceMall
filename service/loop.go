package service

import (
	"bytedancemall/order/pkg"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

func (s *OrderService) LoopUpdateOrderStatus() {
	slog.Info("Start order status update loop")
	for {
		msg, err := pkg.GetReader("order_status").FetchMessage(context.Background())
		if err != nil {
			slog.Error("Failed to fetch message", "error", err)
			continue
		}
		// 处理消息
		err = s.updateOrderStatus(context.Background(), msg)
		if err != nil {
			slog.Error("Failed to update order status", "error", err)
		}
	}
}

type OrderStatus struct {
	OrderID uint32 `json:"order_id"`
	Status  int    `json:"status"`
}

func (s *OrderService) updateOrderStatus(ctx context.Context, msg kafka.Message) error {
	// 处理订单状态更新消息
	var orderStatus OrderStatus
	maxRetries := 5
	err := json.Unmarshal(msg.Value, &orderStatus)
	if err != nil {
		for range maxRetries {
			if err = pkg.GetWriter("order_status_dmq").WriteMessages(ctx, kafka.Message{
				Key:   msg.Key,
				Value: msg.Value,
			}); err == nil {
				return fmt.Errorf("failed to unmarshal order status message, sent to DMQ: %w", err)
			}
		}
		return fmt.Errorf("failed to send to DMQ: %w", err)
	}

	tx := pkg.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("Recovered from panic", "error", r)
		}
	}()

	for range maxRetries {
		if err = tx.Where("order_id = ?", orderStatus.OrderID).Updates(OrderStatus{Status: orderStatus.Status}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update order status: %w", err)
		}
		err = tx.Commit().Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return err
}
