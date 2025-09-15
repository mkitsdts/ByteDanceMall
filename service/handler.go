package service

import (
	"bytedancemall/order/pkg/database"
	"bytedancemall/order/pkg/redis"
	pb "bytedancemall/order/proto"
	"bytedancemall/order/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (s *OrderService) ListOrder(ctx context.Context, req *pb.ListOrderReq) (*pb.ListOrderResp, error) {
	id := req.UserId
	data, err := redis.GetRedisCli().Get(ctx, "order_list_"+fmt.Sprint(id)).Result()
	var orders Orders
	if err == nil {
		if err := json.Unmarshal([]byte(data), &orders); err == nil {
			respBody := make([]*pb.Order, 0, len(orders.Items))
			for _, item := range orders.Items {
				respBody = append(respBody, &pb.Order{
					OrderId:   item.OrderID,
					UserId:    item.UserID,
					ProductId: item.ProductID,
					Amount:    item.Amount,
					Cost:      item.Cost,
					Status:    item.Status,
				})
			}
			return &pb.ListOrderResp{
				Orders: respBody,
			}, nil
		} else {
			return &pb.ListOrderResp{
				Orders: nil,
			}, err
		}
	}
	tx := database.DB().Begin()

	maxRetries := 5
	for i := range maxRetries {
		if err := tx.Where("user_id = ?", id).Find(&orders.Items).Error; err == nil {
			break
		}
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.ListOrderResp{
				Orders: nil,
			}, err
		}
	}
	tx.Commit()
	respBody := make([]*pb.Order, 0, len(orders.Items))
	for _, item := range orders.Items {
		respBody = append(respBody, &pb.Order{
			OrderId:   item.OrderID,
			UserId:    item.UserID,
			ProductId: item.ProductID,
			Amount:    item.Amount,
			Cost:      item.Cost,
			Status:    item.Status,
		})
	}
	return &pb.ListOrderResp{
		Orders: respBody,
	}, nil
}

func (s *OrderService) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusReq) (*pb.GetOrderStatusResp, error) {
	var order Order
	result := redis.GetRedisCli().Get(ctx, "order_status:"+fmt.Sprint(req.OrderId))
	if result.Err() == nil {
		if err := json.Unmarshal([]byte(result.Val()), &order); err == nil {
			return &pb.GetOrderStatusResp{
				OrderStatus: order.Status,
			}, nil
		}
	}
	var err error
	tx := database.DB().Begin()

	maxRetries := 5
	for i := range maxRetries {
		if err = tx.Where("order_id = ?", req.OrderId).First(&order).Error; err == nil {
			return &pb.GetOrderStatusResp{
				OrderStatus: order.Status,
			}, nil
		}
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.GetOrderStatusResp{
				OrderStatus: -1,
			}, err
		}
	}
	tx.Commit()
	return &pb.GetOrderStatusResp{
		OrderStatus: UNKNOWN,
	}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderReq) (*pb.CreateOrderResp, error) {
	tx := database.DB().Begin()

	maxRetries := 5
	var err error
	// 确认订单号和用户ID匹配
	orderKey := "order_id:" + fmt.Sprint(req.OrderId)
	var order_userID uint64
	for i := range maxRetries {
		if res := redis.GetRedisCli().Get(ctx, orderKey); res.Err() == nil {
			order_userID, err = strconv.ParseUint(res.Val(), 10, 64)
			if err != nil {
				return &pb.CreateOrderResp{
					Result: false,
				}, fmt.Errorf("invalid user ID in Redis for order ID")
			}
			break
		}
		if i == maxRetries-1 {
			return &pb.CreateOrderResp{
				Result: false,
			}, fmt.Errorf("order ID not found in Redis after retries")
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	// 不匹配则直接返回
	if order_userID != req.UserId {
		return &pb.CreateOrderResp{
			Result: false,
		}, fmt.Errorf("order ID does not match user ID")
	}

	for i := range maxRetries {
		if err = tx.Create(&Order{
			OrderID:       req.OrderId,
			UserID:        req.UserId,
			ProductID:     req.ProductId,
			Amount:        req.Amount,
			Cost:          req.Cost,
			Status:        WAITING_PAYMENT,
			StreetAddress: req.Address.StreetAddress,
			City:          req.Address.City,
			State:         req.Address.State,
			PaymentStatus: NOT_PAID,
		}).Error; err == nil {
			if err = tx.Commit().Error; err == nil {
				return &pb.CreateOrderResp{
					Result: true,
				}, nil
			}
			time.Sleep(10 << i * time.Millisecond)
		}
		if i == maxRetries-1 {
			tx.Rollback()
			return &pb.CreateOrderResp{
				Result: false,
			}, err
		}
	}
	if err = tx.Commit().Error; err != nil {
		return &pb.CreateOrderResp{
			Result: false,
		}, err
	}

	return &pb.CreateOrderResp{
		Result: false,
	}, nil
}

func (s *OrderService) ApplyOrderID(ctx context.Context, req *pb.ApplyOrderIDReq) (*pb.ApplyOrderIDResp, error) {
	id := utils.GenerateOrderID(req.UserId)
	maxRetries := 5
	for i := range maxRetries {
		if err := redis.GetRedisCli().Set(ctx, "order_id:"+fmt.Sprint(id), req.UserId, 24*time.Hour).Err(); err == nil {
			break
		} else {
			if i == maxRetries-1 {
				return &pb.ApplyOrderIDResp{
					OrderId: 0,
					Result:  false,
				}, err
			}
			time.Sleep(10 << i * time.Millisecond)
		}
	}
	return &pb.ApplyOrderIDResp{
		OrderId: id,
		Result:  true,
	}, nil
}
