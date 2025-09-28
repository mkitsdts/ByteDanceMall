package service

import (
	"bytedancemall/payment/model"
	"bytedancemall/payment/payment"
	"bytedancemall/payment/pkg/database"
	"bytedancemall/payment/pkg/etcd"
	"bytedancemall/payment/pkg/redis"
	pb "bytedancemall/payment/proto"
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
)

func (s *PaymentService) ApplyCharge(ctx context.Context, req *pb.ApplyChargeReq) (*pb.ApplyChargeResp, error) {
	// 先检查 redis 是否存在结果
	order_key := fmt.Sprintf("order_%d", req.OrderId)
	if exists, order_str := redis.CheckDuplicate(order_key, ""); exists && order_str != "failed" {
		return &pb.ApplyChargeResp{
			OrderStr: order_str,
		}, fmt.Errorf("duplicate order")
	}
	maxRetries := 3
	// 如果不存在就尝试获取分布式锁
	order_key_lock := fmt.Sprintf("lock_order_%d", req.OrderId)
	session, err := concurrency.NewSession(etcd.GetEtcdCli(), concurrency.WithTTL(30))
	if err != nil {
		slog.Error("Failed to create etcd session", "error", err)
		return &pb.ApplyChargeResp{
			OrderStr: "",
		}, err
	}
	if session != nil {
		defer session.Close()
	}
	mutex := concurrency.NewMutex(session, order_key_lock)
	if err = mutex.TryLock(ctx); err != nil {
		// 获取锁失败，说明订单正在处理
		time.Sleep(100 * time.Millisecond)
		slog.Info("Order is being processed", "order_id", req.OrderId)
		// 从redis中获取订单号
		for i := range maxRetries {
			if orderStr := redis.GetRedisCli().Get(ctx, order_key); orderStr.Err() == nil {
				if orderStr.Val() != "failed" {
					return &pb.ApplyChargeResp{
						OrderStr: orderStr.Val(),
					}, nil
				} else {
					break
				}
			}
			time.Sleep(50 << i * time.Millisecond)
		}
		// 查找数据库日志，判断订单是否创建
		var records []model.PaymentRecord
		if err := database.DB().Table("payment_records").Where("order_id = ?", req.OrderId).Find(&records).Error; err != nil {
			slog.Error("Failed to query payment create records", "error", err)
			return &pb.ApplyChargeResp{
				OrderStr: "",
			}, err
		}
		if len(records) > 0 {
			for _, record := range records {
				// 处理每个记录
				if record.Status == model.CREATED {
					// 订单创建成功
					go redis.GetRedisCli().Set(ctx, order_key, *record.OrderStr, 5*time.Minute)
					return &pb.ApplyChargeResp{
						OrderStr: *record.OrderStr,
					}, nil
				}
			}
		}
	}

	order_str := s.Client.CreatePaymentRequest(ctx, req.Method, &payment.PaymentRequest{
		OrderID:     req.OrderId,
		UserID:      req.UserId,
		Cost:        req.Cost,
		Description: "Test Payment",
		Attach:      fmt.Sprintf("user_id:%d", req.UserId),
	})

	if order_str == "" {
		err := fmt.Errorf("failed to create payment request")
		// 将订单状态设置为失败，防止重复下单
		redis.GetRedisCli().Set(context.Background(), order_key, "failed", 5*time.Minute)
		// 删除锁

		return &pb.ApplyChargeResp{
			OrderStr: "",
		}, err
	}

	// 成功创建订单后，将订单号存入 redis，防止重复下单
	go redis.GetRedisCli().Set(context.Background(), order_key, order_str, 5*time.Minute)
	return &pb.ApplyChargeResp{
		OrderStr: order_str,
	}, nil
}

func (s *PaymentService) CancelCharge(ctx context.Context, req *pb.CancelChargeReq) (*pb.CancelChargeResp, error) {
	return &pb.CancelChargeResp{}, nil
}
