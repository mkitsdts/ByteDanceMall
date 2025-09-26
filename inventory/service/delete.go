package service

import (
	"bytedancemall/inventory/model"
	pb "bytedancemall/inventory/proto"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (s *InventoryService) DeleteInventory(ctx context.Context, req *pb.DeleteInventoryReq) (*pb.DeleteInventoryResp, error) {
	maxRetries := 10
	for i := range maxRetries {
		_, err := s.Redis.Del(ctx, "product:"+fmt.Sprint(req.ProductId)).Result()
		if err == nil {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			slog.Error("Failed to delete product from Redis")
			return &pb.DeleteInventoryResp{Result: false}, fmt.Errorf("failed to delete product from Redis: %w", err)
		}
	}
	go s.deleteInventoryFromDB(ctx, req.ProductId)
	time.Sleep(300 * time.Millisecond)
	for i := range maxRetries {
		_, err := s.Redis.Del(ctx, "product:"+fmt.Sprint(req.ProductId)).Result()
		if err == nil {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			slog.Error("Failed to delete product from Redis")
			return &pb.DeleteInventoryResp{Result: false}, fmt.Errorf("failed to delete product from Redis: %w", err)
		}
	}
	return &pb.DeleteInventoryResp{Result: true}, nil
}

func (s *InventoryService) deleteInventoryFromDB(ctx context.Context, product_id uint64) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	tx := s.DB.Master.Begin()

	maxRetries := 10
	for i := range maxRetries {
		err := s.DB.Master.Where("product_id = ?", product_id).Delete(&model.Inventory{}).Error
		if err == nil {
			tx.Commit()
			return
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	tx.Rollback()
}
