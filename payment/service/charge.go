package service

import (
	"bytedancemall/payment/model"
	"bytedancemall/payment/payment"
	"bytedancemall/payment/pkg/database"
	"bytedancemall/payment/pkg/etcd"
	"bytedancemall/payment/pkg/redis"
	pb "bytedancemall/payment/proto"
	"bytedancemall/payment/util"
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
	"gorm.io/gorm"
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
			time.Sleep(100 << i * time.Millisecond)
		}
		// 查找数据库日志，判断订单是否创建
		var record model.PaymentOrder
		if err := database.DB().Where(&model.PaymentOrder{}).Where("order_id = ?", req.OrderId).Find(&record).Error; err == nil {
			// 订单已创建
			// 如果支付方式相同且订单已支付，直接返回订单号
			if record.Method == req.Method {
				if record.Status > model.WAITTING {
					go redis.GetRedisCli().Set(ctx, order_key, *record.OrderStr, 5*time.Minute)
					return &pb.ApplyChargeResp{
						OrderStr: *record.OrderStr,
					}, nil
				}
			} else { // 支付方式不同且订单未支付，取消原支付请求，创建新支付请求
				tx := database.DB().Begin()
				var records []model.PaymentRecord
				if err := tx.Model(&model.PaymentRecord{}).Where("order_id = ? and status = ?", req.OrderId, model.CREATED).Find(&records).Error; err != nil {
					slog.Error("Failed to fetch payment records", "error", err)
					tx.Rollback()
					return &pb.ApplyChargeResp{
						OrderStr: "",
					}, fmt.Errorf("unknown database error")
				}
				for _, rec := range records {
					if rec.Status == model.CREATED {
						// 取消原支付请求
						result := make(chan bool, 1)
						go s.Client.CancelPaymentRequest(ctx, record.Method, rec.ID, result)
						// 等待结果
						go func() {
							transaction := database.DB().Begin()
							timer := time.After(10 * time.Second)
							for {
								select {
								case <-timer:
									slog.Error("Timeout waiting for cancel result")
									transaction.Rollback()
									return
								case cancelResult := <-result:
									if cancelResult {
										// 更新订单记录
										if err := transaction.Model(&model.PaymentRecord{}).Where("order_id = ?", req.OrderId).Update("status", model.CANCELED).Error; err != nil {
											slog.Error("Failed to update order status to canceled", "error", err)
											transaction.Rollback()
											return
										}
										transaction.Commit()
									} else {
										slog.Error("Failed to cancel previous payment request")
										transaction.Rollback()
										return
									}
								}
							}
						}()
					}
				}
				// 创建新的支付请求
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
					mutex.Unlock(context.Background())
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
		}

	}
	tx := database.DB().Begin()
	for i := range 3 {
		result := tx.Model(&model.PaymentOrder{}).Where("order_id = ?", req.OrderId).Create(&model.PaymentOrder{
			OrderID: req.OrderId,
			Method:  req.Method,
			Status:  model.CREATED,
		})
		// 这里不能直接用 result.Error 来判断，因为有可能插入成功但是返回的 err 不为 nil
		if err = result.Error; err != nil && err != gorm.ErrDuplicatedKey {
			slog.Error("Failed to create payment record, retrying...", "error", err)
			time.Sleep(10 << i * time.Millisecond)
		}
	}

	uuid := util.GenerateUUID()

	order_str := s.Client.CreatePaymentRequest(ctx, req.Method, &payment.PaymentRequest{
		ID:          uuid,
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
		mutex.Unlock(context.Background())
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
