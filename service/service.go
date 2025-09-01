package service

import (
	pb "bytedancemall/order/proto"

	"github.com/go-redsync/redsync/v4"
)

type OrderService struct {
	Redsync *redsync.Redsync
	pb.UnimplementedOrderServiceServer
}

// 创建一个新的订单服务实例
func NewOrderService() *OrderService {
	var s OrderService
	s.LoopUpdateOrderStatus()
	return &s
}
