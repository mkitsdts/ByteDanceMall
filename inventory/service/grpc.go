package service

import (
	pb "bytedancemall/inventory/proto"
	"context"
)

func (s *InventoryService) DeductInventory(ctx context.Context, req *pb.DeductInventoryReq) (*pb.DeductInventoryResp, error) {
	record, err := s.usecase.Deduct(ctx, req.GetProduct().GetProductId(), req.GetRecordId(), req.GetProduct().GetAmount())
	if err != nil {
		return &pb.DeductInventoryResp{Result: false}, err
	}
	return &pb.DeductInventoryResp{
		Result:   true,
		RecordId: record.RecordID,
	}, nil
}

func (s *InventoryService) CreateInventory(ctx context.Context, req *pb.CreateInventoryReq) (*pb.CreateInventoryResp, error) {
	if err := s.usecase.Create(ctx, req.GetProductId(), req.GetAmount()); err != nil {
		return &pb.CreateInventoryResp{Result: false}, err
	}
	return &pb.CreateInventoryResp{Result: true}, nil
}

func (s *InventoryService) DeleteInventory(ctx context.Context, req *pb.DeleteInventoryReq) (*pb.DeleteInventoryResp, error) {
	if err := s.usecase.Delete(ctx, req.GetProductId()); err != nil {
		return &pb.DeleteInventoryResp{Result: false}, err
	}
	return &pb.DeleteInventoryResp{Result: true}, nil
}

func (s *InventoryService) QueryInventory(ctx context.Context, req *pb.QueryInventoryReq) (*pb.QueryInventoryResp, error) {
	inventory, err := s.usecase.Query(ctx, req.GetProductId())
	if err != nil {
		return nil, err
	}
	return &pb.QueryInventoryResp{
		ProductId:      inventory.ProductID,
		AvailableStock: availableStock(inventory.TotalStock, inventory.LockedStock),
	}, nil
}

func (s *InventoryService) PreheatInventory(ctx context.Context, req *pb.PreheatInventoryReq) (*pb.PreheatInventoryResp, error) {
	productIDs := make([]uint64, 0, len(req.GetProducts()))
	for _, product := range req.GetProducts() {
		productIDs = append(productIDs, product.GetProductId())
	}
	if err := s.usecase.Preheat(ctx, productIDs); err != nil {
		return &pb.PreheatInventoryResp{Result: false}, err
	}
	return &pb.PreheatInventoryResp{Result: true}, nil
}
