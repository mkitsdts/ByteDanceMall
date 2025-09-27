package service

import (
	"bytedancemall/payment/pkg/redis"
	pb "bytedancemall/payment/proto"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (s *PaymentService) ApplyCharge(ctx context.Context, req *pb.ApplyChargeReq) (*pb.ApplyChargeResp, error) {
	// 先检查 redis 是否存在重复订单
	order_key := fmt.Sprintf("order_%d", req.OrderId)
	if exists, order_str := redis.CheckDuplicate(order_key); exists {
		return &pb.ApplyChargeResp{
			OrderStr: order_str,
		}, fmt.Errorf("duplicate order")
	}

	// 如果不存在就尝试获取分布式锁
	order_key_lock := fmt.Sprintf("lock_order_%d", req.OrderId)
	if exists, method := redis.CheckDuplicate(order_key_lock); exists {
		if method == getPaymentMethodName(req.Method) {
			return &pb.ApplyChargeResp{
				OrderStr: "",
			}, fmt.Errorf("order is being processed")
		}
	}

	// 如果不存在就尝试获取锁
	maxRetries := 5
	for i := range maxRetries {
		if redis.GetRedisCli().SetNX(context.Background(), order_key_lock, getPaymentMethodName(req.Method), 5*time.Second).Val() {
			break
		}
		time.Sleep(10 << i * time.Millisecond)
		if i == maxRetries-1 {
			slog.Error("!!!failed to acquire lock", "order_id", req.OrderId)
			return &pb.ApplyChargeResp{
				OrderStr: "",
			}, fmt.Errorf("order is being processed")
		}
	}

	var order_str string
	var err error
	if req.Method == WECHAT_PAY {
		order_str, err = s.wechat_pay(ctx, req.Cost, req.OrderId, req.UserId)
	} else if req.Method == ALIPAY {
		order_str, err = s.alipay(ctx, req.Cost, req.OrderId, req.UserId)
	}
	if err != nil {
		slog.Error("failed to create order", "error", err)
		return nil, err
	}
	return &pb.ApplyChargeResp{
		OrderStr: order_str,
	}, nil
}

func (s *PaymentService) CancelCharge(ctx context.Context, req *pb.CancelChargeReq) (*pb.CancelChargeResp, error) {
	return &pb.CancelChargeResp{}, nil
}

func (s *PaymentService) wechat_pay(ctx context.Context, cost int64, OrderId, UserId uint64) (string, error) {
	return "114514", nil
}

func (s *PaymentService) alipay(ctx context.Context, cost int64, OrderId, UserId uint64) (string, error) {
	return "114514", nil
}
