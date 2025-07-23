package service

import (
	pb "bytedancemall/order/proto"
	"context"
	"encoding/json"
	"strconv"
	"time"
)

func (s *OrderService) ListOrder(ctx context.Context, req *pb.ListOrderReq) (*pb.ListOrderResp, error) {
	// 从Redis中获取订单
	val := s.Redis.Get(ctx, strconv.FormatUint(uint64(req.UserId), 10)).Val()
	if val != "" {
		// 返回订单
		var pbOrders []*pb.Order
		orders := []Order{}
		json.Unmarshal([]byte(val), &orders)
		for _, order := range orders {
			pbOrders = append(pbOrders, &pb.Order{
				OrderId: order.OrderId,
				UserId:  order.UserId,
				Address: &pb.Address{
					StreetAddress: order.StreetAddress,
					City:          order.City,
					State:         order.State,
				},
			})
		}
		return &pb.ListOrderResp{Orders: pbOrders}, nil
	}

	// redis中没有订单，从数据库中获取订单
	var orders []Order
	s.Db.Master.Where("user_id = ?", req.UserId).Find(&orders)
	// 返回订单
	var pbOrders []*pb.Order
	for _, order := range orders {
		pbOrders = append(pbOrders, &pb.Order{
			OrderId: order.OrderId,
			UserId:  order.UserId,
			Address: &pb.Address{
				StreetAddress: order.StreetAddress,
				City:          order.City,
				State:         order.State,
			},
		})
	}

	// 将订单存入Redis
	bytes, _ := json.Marshal(orders)
	s.Redis.Set(ctx, strconv.FormatUint(uint64(req.UserId), 10), bytes, 5*time.Minute)
	return &pb.ListOrderResp{Orders: pbOrders}, nil
}

func (s *OrderService) GetOrderStates(ctx context.Context, req *pb.GetOrderStatesReq) (*pb.GetOrderStatesResp, error) {
	// 从Redis中获取订单状态
	val := s.Redis.Get(ctx, strconv.FormatUint(req.OrderId, 10)).Val()
	if val != "" {
		return &pb.GetOrderStatesResp{OrderStatuses: pb.OrderStatus_ORDER_STATUS_PENDING}, nil
	}

	// redis中没有订单状态，从数据库中获取订单状态
	var order Order
	if err := s.Db.Master.Where("order_id = ?", req.OrderId).First(&order).Error; err != nil {
		return nil, err
	}

	var result pb.OrderStatus
	switch order.State {
	case "WAITING_PAYMENT":
		order.State = pb.OrderStatus_ORDER_STATUS_PENDING.String()
		result = pb.OrderStatus_ORDER_STATUS_PENDING
	case "PAID":
		order.State = pb.OrderStatus_ORDER_STATUS_PAID.String()
		result = pb.OrderStatus_ORDER_STATUS_PAID
	case "SHIPPED":
		order.State = pb.OrderStatus_ORDER_STATUS_SHIPPED.String()
		order.State = pb.OrderStatus_ORDER_STATUS_SHIPPED.String()
	case "DELIVERED":
		order.State = pb.OrderStatus_ORDER_STATUS_DELIVERED.String()
		order.State = pb.OrderStatus_ORDER_STATUS_DELIVERED.String()
	case "CANCELED":
		order.State = pb.OrderStatus_ORDER_STATUS_CANCELED.String()
		order.State = pb.OrderStatus_ORDER_STATUS_CANCELED.String()
	default:
		order.State = pb.OrderStatus_ORDER_STATUS_UNSPECIFIED.String()
		result = pb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}

	// 将订单状态存入Redis
	s.Redis.Set(ctx, strconv.FormatUint(req.OrderId, 10), order.State, 5*time.Minute)
	return &pb.GetOrderStatesResp{OrderStatuses: result}, nil
}
