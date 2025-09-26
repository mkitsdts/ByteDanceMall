package service

import (
	"bytedancemall/inventory/model"
	pb "bytedancemall/inventory/proto"
	"context"
	"fmt"
	"log/slog"
	"time"
)

// 预扣减库存
func (s *InventoryService) DeductInventory(ctx context.Context, req *pb.DeductInventoryReq) (*pb.DeductInventoryResp, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	maxRetries := 3

	lockKey := fmt.Sprint("lock:order:", req.OrderId)
	// 尝试获取分布式锁
waitLoop:
	for i := range maxRetries {
		result := s.Redis.SetNX(ctx, lockKey, "0", 30*time.Second)
		if result.Val() {
			// 成功获取锁，确保在函数退出时释放锁
			defer s.Redis.Del(ctx, lockKey)
			break
		} else if result.Err() == nil {
			// 等待扣减结果
			r := s.Redis.Get(ctx, lockKey)
			if r.Err() != nil && r.Err() != context.Canceled && r.Err() != context.DeadlineExceeded {
				slog.Error("Failed to get Redis lock status", "error", r.Err())
				return &pb.DeductInventoryResp{
					Result: false,
				}, r.Err()
			}
			// 如果锁状态是1，表示扣减成功；如果是2，表示扣减失败
			if r.Val() == "1" {
				return &pb.DeductInventoryResp{
					Result: true,
				}, nil
			}
			if r.Val() == "2" {
				return &pb.DeductInventoryResp{
					Result: false,
				}, nil
			}
			for {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(10 * time.Second):
					break waitLoop
				default:
					time.Sleep(50 * time.Millisecond)
					result := s.Redis.Get(ctx, lockKey)
					if result.Err() == nil {
						if result.Val() == "1" {
							return &pb.DeductInventoryResp{
								Result: true,
							}, nil
						} else if result.Val() == "2" {
							return &pb.DeductInventoryResp{
								Result: false,
							}, nil
						}
					}
				}
			}
		}
		if i == maxRetries-1 {
			slog.Error("Failed to execute Redis script", "error", result.Err())
			return &pb.DeductInventoryResp{
				Result: false,
			}, result.Err()
		}
		time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
	}

	if state := s.inventoryExist(req.OrderId); state == 1 {
		return &pb.DeductInventoryResp{
			Result: true,
		}, nil
	} else if state == 2 {
		return &pb.DeductInventoryResp{
			Result: false,
		}, nil
	}

	// 扣减库存
	if err := s.deduct(req.Product.ProductId, req.Product.Amount, req.OrderId); err != nil {
		slog.Error("Failed to deduct inventory", "error", err, "product_id", req.Product.ProductId, "amount", req.Product.Amount)
		// 扣减失败，设置锁状态为失败
		result := s.Redis.Set(ctx, lockKey, "2", 30*time.Second)
		if result.Err() == nil {
			return &pb.DeductInventoryResp{
				Result: true,
			}, nil
		}
		return &pb.DeductInventoryResp{
			Result: false,
		}, err
	}
	result := s.Redis.Set(ctx, lockKey, "1", 30*time.Second)
	if result.Err() == nil {
		return &pb.DeductInventoryResp{
			Result: true,
		}, nil
	}
	return &pb.DeductInventoryResp{
		Result: false,
	}, nil
}

func (s *InventoryService) deduct(product_id uint64, amount uint64, order_id uint64) error {
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
		if inventory.TotalStock-inventory.LockedStock < amount {
			tx.Rollback()
			return fmt.Errorf("insufficient inventory for product_id %d: available %d, requested %d", product_id, inventory.TotalStock, amount)
		}
		// 乐观锁更新库存
		result := tx.Model(&inventory).
			Where("product_id = ? AND version = ?", product_id, inventory.Version).
			Updates(map[string]any{
				"locked_stock": inventory.LockedStock + amount,
				"version":      inventory.Version + 1,
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

		// 记录出库信息
		outInventory := model.OutInventory{
			ProductID: product_id,
			OrderId:   order_id,
			Amount:    amount,
			State:     1, // 1表示预扣减成功
		}
		if err := tx.Create(&outInventory).Error; err != nil {
			tx.Rollback()
			slog.Error("Failed to create out inventory record", "error", err)
			return err
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			return err
		}
	}
	return nil
}

func (s *InventoryService) inventoryExist(OrderID uint64) int8 {
	state := int8(-128)
	tx := s.DB.Master.Begin()
	result := tx.Select("State").Where("order_id = ?", OrderID).First(&state)
	if result.Error != nil {
		tx.Rollback()
		return -128
	}
	tx.Commit()
	return state
}
