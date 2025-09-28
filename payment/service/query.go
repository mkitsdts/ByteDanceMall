package service

import (
	"bytedancemall/payment/model"
	pb "bytedancemall/payment/proto"
	"context"
)

func (s *PaymentService) QueryStatus(ctx context.Context, req *pb.QueryStatusReq) (*pb.QueryStatusResp, error) {
	return &pb.QueryStatusResp{
		OrderId: req.OrderId,
		Status:  model.WAITTING,
	}, nil
}
