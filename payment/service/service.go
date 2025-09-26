package service

import (
	pb "bytedancemall/payment/proto"
)

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
}

// 创建一个新的用户服务实例
func NewPaymentService() *PaymentService {
	var service PaymentService
	return &service
}
