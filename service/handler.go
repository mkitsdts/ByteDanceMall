package service

import (
	pb "bytedancemall/cart/proto"
	"context"
)

// 添加商品至购物车
func (s *CartService) AddItem(ctx context.Context, req *pb.AddItemReq) (*pb.AddItemResp, error) {
	return &pb.AddItemResp{}, nil
}

// 获取购物车中的商品
func (s *CartService) GetCart(ctx context.Context, req *pb.GetCartReq) (*pb.GetCartResp, error) {
	return &pb.GetCartResp{}, nil
}

func (s *CartService) RemoveItem(ctx context.Context, req *pb.RemoveItemReq) (*pb.RemoveItemResp, error) {
	return &pb.RemoveItemResp{}, nil
}
