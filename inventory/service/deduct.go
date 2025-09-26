package service

import (
	"bytedancemall/inventory/model"
	pb "bytedancemall/inventory/proto"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// 预扣减库存
func (s *InventoryService) DeductInventory(ctx context.Context, req *pb.DeductInventoryReq) (*pb.DeductInventoryResp, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	maxRetries := 3

	var result any
	var err error
	for i := range maxRetries {
		result, err = s.deductScript.Run(
			ctx,
			s.Redis,
			[]string{fmt.Sprintf("product:%d", req.Product.ProductId)},
			req.Product.Amount,
		).Result()
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			slog.Error("Failed to execute Redis script", "error", err)
			return &pb.DeductInventoryResp{
				Result: false,
			}, err
		}
	}

	switch v := result.(type) {
	case int64:
		// Redis 不存在该库存消息
		if v == -1 {
			slog.Warn("Inventory not found in Redis", "product_id", req.Product.ProductId)
			var inventory model.Inventory
			if err := s.DB.Master.Where("product_id = ?", req.Product.ProductId).First(&inventory).Error; err != nil {
				slog.Error("Failed to find inventory in DB", "error", err)
				if err == gorm.ErrRecordNotFound {
					go s.Redis.Set(ctx, fmt.Sprintf("product:%d", req.Product.ProductId), 0, 0)
				}
				return &pb.DeductInventoryResp{
					Result: false,
				}, err
			}
			return &pb.DeductInventoryResp{
				Result: false,
			}, nil
		}
		if v == -2 {
			slog.Info("Insufficient inventory in Redis", "product_id", req.Product.ProductId)
			return &pb.DeductInventoryResp{
				Result: false,
			}, nil
		}
		// 异步发送成功消息
		go s.writeDeductSuccessMsg(req)
		return &pb.DeductInventoryResp{
			Result: true,
		}, nil
	default:
		slog.Error("Unexpected result type from Redis script", "type", fmt.Sprintf("%T", v))
		return &pb.DeductInventoryResp{
			Result: false,
		}, fmt.Errorf("unexpected result type from Redis script: %T", v)
	}

}

func (s *InventoryService) writeDeductSuccessMsg(req *pb.DeductInventoryReq) {
	msg := model.DeductMessage{
		ProductId: req.Product.ProductId,
		Amount:    req.Product.Amount,
	}
	body, _ := json.Marshal(msg)
	for i := range 3 {
		if err := s.Writer["gomall-inventory-deduct"].WriteMessages(context.Background(), kafka.Message{
			Key:   fmt.Appendf(nil, "%d", req.OrderId),
			Value: body,
		}); err == nil {
			slog.Info("Successfully wrote deduct message to Kafka", "inventory_id", req.Product.ProductId, "amount", req.Product.Amount)
			return
		}
		// 指数退避重试
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	slog.Error("Failed to write message to Kafka after retries", "inventory_id", req.Product.ProductId)
}

func (s *InventoryService) deduct(product_id uint64, amount uint64) error {
	slog.Info("Starting inventory deduction", "product_id", product_id, "amount", amount)
	// 乐观锁重试机制
	maxRetries := 10
	for retry := range maxRetries {
		tx := s.DB.Master.Begin()
		var inventory model.Inventory
		if err := tx.Where("product_id = ?", product_id).First(&inventory).Error; err != nil {
			tx.Rollback()
			return err
		}
		// 检查库存
		if inventory.TotalStock < amount {
			tx.Rollback()
			return fmt.Errorf("insufficient inventory for product_id %d: available %d, requested %d", product_id, inventory.TotalStock, amount)
		}
		// 乐观锁更新库存
		result := tx.Model(&inventory).
			Where("product_id = ? AND version = ?", product_id, inventory.Version).
			Updates(map[string]any{
				"locked_stock": inventory.LockedStock + amount,
				"version":      inventory.Version + 1,
				"updated_at":   time.Now(),
			})

		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		// 乐观锁冲突检测
		if result.RowsAffected == 0 {
			tx.Rollback()
			slog.Warn("Optimistic lock conflict, retrying", "retry", retry+1, "inventory_id", product_id)

			if retry == maxRetries-1 {
				slog.Error("Max retries reached for optimistic lock conflict", "inventory_id", product_id)
				return fmt.Errorf("optimistic lock conflict for product_id %d", product_id)
			}

			time.Sleep(time.Duration(1<<retry) * 10 * time.Millisecond)
			continue
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			return err
		}
	}
	return nil
}
