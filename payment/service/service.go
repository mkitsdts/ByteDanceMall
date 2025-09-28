package service

import (
	"bytedancemall/payment/payment"
	pb "bytedancemall/payment/proto"
)

type PaymentService struct {
	Client *payment.PaymentClient
	pb.UnimplementedPaymentServiceServer
}

// 创建一个新的用户服务实例
func NewPaymentService() *PaymentService {
	var service PaymentService
	service.Client = payment.NewPaymentClient()
	return &service
}
