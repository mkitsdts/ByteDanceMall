package service

import (
	"bytedancemall/inventory/model"
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func (s *InventoryService) ConsumeReduceMsg(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		for {
			msg, err := s.Reader["reduce"].ReadMessage(ctx)
			if err != nil {
				slog.Error("Failed to read message", "error", err)
				continue
			}
			slog.Info("Received reduce message", "key", string(msg.Key), "value", string(msg.Value))
			// 处理消息
			go s.reduce(ctx, msg)
		}
	}
}

// 扣减库存
func (s *InventoryService) reduce(ctx context.Context, msg kafka.Message) {
	select {
	case <-ctx.Done():
		slog.Warn("Context done, stopping reduce operation")
		return
	default:
	}
	// 持久化至数据库
	slog.Info("Reducing inventory", "key", string(msg.Key), "value", string(msg.Value))

	tx := s.DB.Master.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("Transaction rolled back due to panic", "error", r)
		}
	}()
	var inventory model.Inventory
	var valueUint uint64
	if err := tx.Where("id = ?", string(msg.Key)).First(&inventory).Error; err != nil {
		slog.Error("Failed to find inventory", "error", err)
		valueStr := string(msg.Value)
		valueUint, err = strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			slog.Error("Failed to parse value to uint64", "error", err)
			tx.Rollback()
			return
		}
		if inventory.TotalStock < valueUint {
			slog.Warn("Insufficient inventory", "key", string(msg.Key), "available", inventory.TotalStock, "requested", valueUint)
			tx.Rollback()
			return
		}
		return
	}
	inventory.TotalStock -= valueUint
	inventory.LockedStock += valueUint
	inventory.Version++
	inventory.UpdatedAt = time.Now()
	if err := tx.Save(&inventory).Error; err != nil {
		slog.Error("Failed to update inventory", "error", err)
		tx.Rollback()
		return
	}
	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		tx.Rollback()
		return
	}
	slog.Info("Successfully reduced inventory", "key", string(msg.Key), "new_amount", inventory.TotalStock)
	s.Reader["reduce"].CommitMessages(ctx, msg)
	s.Redis.Del(ctx, "inventory:"+string(msg.Key)) // 删除缓存
	time.Sleep(100 * time.Millisecond)
	slog.Info("Cache cleared for inventory", "key", string(msg.Key))
	s.Redis.Del(ctx, "inventory:"+string(msg.Key))
}
