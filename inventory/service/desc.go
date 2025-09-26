package service

import (
	"bytedancemall/inventory/model"
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func (s *InventoryService) LoopReduce() {
	slog.Info("Starting LoopReduce...")
	for {
		msg, err := s.Reader["gomall-inventory-reduce"].FetchMessage(context.Background())
		if err != nil {
			slog.Error("Failed to read message", "error", err)
			continue
		}
		slog.Info("Received reduce message", "key", string(msg.Key), "value", string(msg.Value))
		// 处理消息
		go s.reduce(msg)
		time.Sleep(50 * time.Millisecond)
	}
}

// 扣减库存
func (s *InventoryService) reduce(msg kafka.Message) error {
	// 持久化至数据库
	slog.Info("Reducing inventory", "key", string(msg.Key), "value", string(msg.Value))
	ctx := context.Background()
	tx := s.DB.Master.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("Transaction rolled back due to panic", "error", r)
		}
	}()
	var inventory model.Inventory
	var valueUint uint64
	if err := tx.Where("product_id = ?", string(msg.Key)).First(&inventory).Error; err != nil {
		slog.Error("Failed to find inventory", "error", err)
		valueStr := string(msg.Value)
		valueUint, err = strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			slog.Error("Failed to parse value to uint64", "error", err)
			tx.Rollback()
			return err
		}
		if inventory.TotalStock < valueUint {
			slog.Warn("Insufficient inventory", "key", string(msg.Key), "available", inventory.TotalStock, "requested", valueUint)
			tx.Rollback()
			return err
		}
		return nil
	}
	inventory.TotalStock -= valueUint
	inventory.LockedStock -= valueUint
	inventory.Version++
	inventory.UpdatedAt = time.Now()
	if err := tx.Save(&inventory).Error; err != nil {
		slog.Error("Failed to update inventory", "error", err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		tx.Rollback()
		return err
	}
	slog.Info("Successfully reduced inventory", "key", string(msg.Key), "new_amount", inventory.TotalStock)
	s.Reader["gomall-inventory-reduce"].CommitMessages(ctx, msg)
	// 延迟双删
	s.Redis.Del(ctx, "inventory:"+string(msg.Key)) // 删除缓存
	time.Sleep(100 * time.Millisecond)
	slog.Info("Cache cleared for inventory", "key", string(msg.Key))
	s.Redis.Del(ctx, "inventory:"+string(msg.Key))
	return nil
}
