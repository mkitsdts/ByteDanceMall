package service

import (
	pb "bytedancemall/seckill/proto"
	"context"
	"fmt"
	"log/slog"
)

const layout = "2006-01-02 15:04:05"

func (s *SeckillSer) AddItem(ctx context.Context, req *pb.AddItemReq) (*pb.AddItemResp, error) {
	slog.Info("AddItem", "request", req)

	// 检查请求参数
	if req.ProductId == 0 || req.Quantity == 0 || req.ReleaseTime == "" {
		return nil, fmt.Errorf("invalid request parameters")
	}

	// 添加商品到秒杀列表
	s.AddItemHandler(ctx, req.ProductId, req.Quantity, req.ReleaseTime)

	// 响应成功
	return &pb.AddItemResp{
		Result: true,
	}, nil
}

func (s *SeckillSer) TrySecKillItem(ctx context.Context, req *pb.TrySecKillItemReq) (*pb.TrySecKillItemResp, error) {
	slog.Info("TrySecKillItem", "request", req)

	// 检查请求参数
	if req.ProductId == 0 || req.UserId == 0 {
		return nil, fmt.Errorf("invalid request parameters")
	}

	// 尝试秒杀商品
	result := s.TrySecKillItemHandler(ctx, req.ProductId, req.UserId)
	if result != nil {
		slog.Error("TrySecKillItem failed", "product_id", req.ProductId, "user_id", req.UserId, "result", result)
		return nil, fmt.Errorf("failed to try sec kill item, result code: %d", result)
	}

	// 响应结果
	return &pb.TrySecKillItemResp{
		Result: 1,
	}, nil
}
