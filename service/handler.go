package service

import (
	"bytedancemall/order/pkg"
	pb "bytedancemall/order/proto"
	"bytedancemall/order/utils"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

func (s *OrderService) ListOrder(ctx context.Context, req *pb.ListOrderReq) (*pb.ListOrderResp, error) {
	id := req.UserId
	data, err := pkg.GetRedisCli().Get(ctx, "order_list_"+fmt.Sprint(id)).Result()
	var orders Orders
	if err == nil {
		if err := json.Unmarshal([]byte(data), &orders); err == nil {
			respBody := make([]*pb.Order, 0, len(orders.Items))
			for _, item := range orders.Items {
				respBody = append(respBody, &pb.Order{
					OrderId:   item.OrderId,
					UserId:    item.UserId,
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
	tx := pkg.DB().Begin()

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
			OrderId:   item.OrderId,
			UserId:    item.UserId,
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
	result, err := pkg.GetRedisCli().Get(ctx, "order_"+fmt.Sprint(req.OrderId)).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(result), &order); err == nil {
			return &pb.GetOrderStatusResp{
				OrderStatus: order.Status,
			}, nil
		}
	}
	tx := pkg.DB().Begin()

	maxRetries := 5
	for i := range maxRetries {
		if err := tx.Where("order_id = ?", req.OrderId).First(&order).Error; err == nil {
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
		OrderStatus: -1,
	}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderReq) (*pb.CreateOrderResp, error) {
	tx := pkg.DB().Begin()

	maxRetries := 5
	var err error
	for i := range maxRetries {
		if err = tx.Create(&Order{
			UserId:    req.UserId,
			ProductID: req.ProductId,
			Amount:    req.Amount,
			Cost:      req.Cost,
			Status:    WAITING_PAYMENT,
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
	go func() {
		time.AfterFunc(5*time.Minute, func() {
			for i := range maxRetries {
				if err := pkg.DB().Model(&Order{}).Where("order_id = ? and status = ?", req.OrderId, WAITING_PAYMENT).Update("status", CANCELED).Error; err == nil {
					return
				}
				time.Sleep(10 << i * time.Millisecond)
				if i == maxRetries-1 {
					slog.Error("Failed to auto-cancel order after retries", "order_id", req.OrderId)
				}
			}
		})
	}()
	return &pb.CreateOrderResp{
		Result: true,
	}, nil
}

func ApplyOrderID(ctx context.Context, req *pb.ApplyOrderIDReq) (*pb.ApplyOrderIDResp, error) {
	id := utils.GenerateOrderID(req.UserId)
	return &pb.ApplyOrderIDResp{
		OrderId: id,
		Result:  true,
	}, nil
}
