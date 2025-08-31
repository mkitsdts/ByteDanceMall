package service

import (
	"bytedancemall/inventory/model"
	pb "bytedancemall/inventory/proto"
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func (s *InventoryService) RecoverInventory(ctx context.Context, req *pb.RecoverInventoryReq) (*pb.RecoverInventoryResp, error) {
	go s.Recover(ctx, req.ProductId, req.Amount)
	return &pb.RecoverInventoryResp{
		Result: true,
	}, nil
}

func (s *InventoryService) Recover(ctx context.Context, product_id uint64, amount uint64) {
	tx := s.DB.Master.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("Transaction rolled back due to panic", "error", r)
		}
	}()
	var inventory model.Inventory
	if err := tx.Where("product_id = ?", product_id).First(&inventory).Error; err != nil {
		slog.Error("Failed to find inventory", "error", err)
	}
	if inventory.TotalStock < amount {
		slog.Warn("Insufficient inventory", "id", product_id, "available", inventory.TotalStock, "requested", amount)
		tx.Rollback()
		return
	}
	inventory.TotalStock += amount
	inventory.LockedStock -= amount
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
	slog.Info("Successfully reduced inventory", "id", product_id, "new_amount", inventory.TotalStock)
}

func (s *InventoryService) ConsumeRecoverMsgLoop() {
	for {
		msg, err := s.Reader["gomall-inventory-recover"].FetchMessage(context.Background())
		if err != nil {
			slog.Error("Failed to read message", "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		slog.Info("Received recover message", "key", string(msg.Key), "value", string(msg.Value))
		// 处理消息
		go s.ConsumeRecoverMsg(context.Background(), msg)
	}
}

func (s *InventoryService) ConsumeRecoverMsg(ctx context.Context, msg kafka.Message) {
	// 持久化至数据库
	slog.Info("Recovering inventory", "key", string(msg.Key), "value", string(msg.Value))

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
			return
		}
		if inventory.TotalStock < valueUint {
			slog.Warn("Insufficient inventory", "key", string(msg.Key), "available", inventory.TotalStock, "requested", valueUint)
			tx.Rollback()
			return
		}
		return
	}
	inventory.TotalStock += valueUint
	inventory.LockedStock -= valueUint
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
	s.Reader["gomall-inventory-recover"].CommitMessages(ctx, msg)
	s.Redis.Del(ctx, "inventory:"+string(msg.Key)) // 删除缓存
	time.Sleep(100 * time.Millisecond)
	slog.Info("Cache cleared for inventory", "key", string(msg.Key))
	s.Redis.Del(ctx, "inventory:"+string(msg.Key)) // 再次删除缓存以确保一致性
}
