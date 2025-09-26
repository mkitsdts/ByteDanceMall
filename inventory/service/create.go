package service

import (
	"bytedancemall/inventory/model"
	pb "bytedancemall/inventory/proto"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (s *InventoryService) CreateInventory(ctx context.Context, req *pb.CreateInventoryReq) (*pb.CreateInventoryResp, error) {
	inventory := model.Inventory{
		ProductID:   req.ProductId,
		TotalStock:  req.Amount,
		LockedStock: 0,
		State:       model.StateOnSale,
	}

	tx := s.DB.Master.Begin()
	if err := tx.Create(&inventory).Error; err != nil {
		slog.Error("Failed to create inventory", "error", err)
		tx.Rollback()
		return &pb.CreateInventoryResp{
			Result: false,
		}, err
	}

	tx.Commit()
	go s.syncToRedis(inventory)
	return &pb.CreateInventoryResp{
		Result: true,
	}, nil
}

func (s *InventoryService) syncToRedis(inventory model.Inventory) {
	key := "product:" + fmt.Sprintf("%d", inventory.ProductID)

	maxRetries := 10
	for range maxRetries {
		if err := s.Redis.Set(context.Background(), key, inventory.TotalStock, 0).Err(); err != nil {
			slog.Error("Failed to sync inventory to Redis", "error", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		slog.Info("Successfully synced inventory to Redis", "product_id", inventory.ProductID, "amount", inventory.TotalStock)
		return
	}
	slog.Error("Exceeded max retries to sync inventory to Redis", "product_id", inventory.ProductID)
}
