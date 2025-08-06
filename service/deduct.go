package service

import (
	"bytedancemall/inventory/model"
	pb "bytedancemall/inventory/proto"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

func (s *InventoryService) DeductInventory(ctx context.Context, req *pb.DeduatInventoryReq) (*pb.DeduatInventoryResp, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 乐观锁重试机制
	maxRetries := 3
	for retry := range maxRetries {
		tx := s.DB.Master.Begin()

		var inventory model.Inventory
		if err := tx.Where("id = ?", req.InventoryId).First(&inventory).Error; err != nil {
			tx.Rollback()
			slog.Error("Failed to find inventory", "error", err)
			return nil, fmt.Errorf("failed to find inventory: %w", err)
		}

		// 检查库存
		if inventory.TotalStock < req.Amount {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient stock: available %d, requested %d", inventory.TotalStock, req.Amount)
		}

		// 乐观锁更新库存
		result := tx.Model(&inventory).
			Where("id = ? AND version = ?", req.InventoryId, inventory.Version).
			Updates(map[string]any{
				"stock":      inventory.TotalStock - req.Amount,
				"version":    inventory.Version + 1,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			tx.Rollback()
			slog.Error("Failed to update inventory", "error", result.Error)
			return nil, fmt.Errorf("failed to update inventory: %w", result.Error)
		}

		// 乐观锁冲突检测
		if result.RowsAffected == 0 {
			tx.Rollback()
			slog.Warn("Optimistic lock conflict, retrying", "retry", retry+1, "inventory_id", req.InventoryId)

			if retry == maxRetries-1 {
				return nil, fmt.Errorf("optimistic lock conflict after %d retries", maxRetries)
			}

			time.Sleep(time.Duration(1<<retry) * 10 * time.Millisecond)
			continue
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		// 异步发送成功消息
		go s.writeDeductSuccessMsg(ctx, req)
		go s.writeDeductSuccessToRedis(ctx, inventory)
		return &pb.DeduatInventoryResp{
			Result: true,
		}, nil
	}

	return nil, fmt.Errorf("failed to deduct inventory after %d retries", maxRetries)
}

func (s *InventoryService) writeDeductSuccessMsg(ctx context.Context, req *pb.DeduatInventoryReq) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	for i := range 3 {
		if err := s.Writer.WriteMessages(ctx, kafka.Message{
			Key:   fmt.Appendf(nil, "%d", req.InventoryId),
			Value: fmt.Appendf(nil, "Deducted %d from inventory %d", req.Amount, req.InventoryId),
		}); err != nil {
			return
		}
		// 指数退避重试
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	slog.Error("Failed to write message to Kafka after retries", "inventory_id", req.InventoryId)
}

func (s *InventoryService) writeDeductSuccessToRedis(ctx context.Context, inventory model.Inventory) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	for i := range 3 {
		if err := s.Redis.Set(ctx, fmt.Sprintf("deduct_inventory:%d", inventory.ProductID), inventory.TotalStock, 0).Err(); err != nil {
			return
		}
		// 指数退避重试
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	slog.Error("Failed to write message to Redis after retries", "inventory_id", inventory.ProductID)
}
