package service

import (
	pb "bytedancemall/payment/proto"
	"context"
)

func (s *PaymentService) QueryStatus(ctx context.Context, req *pb.QueryStatusReq) (*pb.QueryStatusResp, error) {
	return &pb.QueryStatusResp{
		OrderId: req.OrderId,
		Status:  WAITTING,
	}, nil
}
