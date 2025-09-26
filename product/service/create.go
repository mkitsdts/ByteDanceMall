package service

import (
	"bytedancemall/product/model"
	"bytedancemall/product/pkg/database"
	pb "bytedancemall/product/proto"
	"context"
	"fmt"
	"log/slog"
)

func (s *ProductService) CreateProducts(ctx context.Context, req *pb.CreateProductsReq) (*pb.CreateProductsResp, error) {
	if req.Product == nil {
		return &pb.CreateProductsResp{
			Result: false,
		}, fmt.Errorf("product cannot be nil")
	}
	select {
	case <-ctx.Done():
		return &pb.CreateProductsResp{
			Result: false,
		}, ctx.Err()
	default:
	}

	for _, product := range req.Product {
		newProduct := model.Product{
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		}
		result := database.Get().Create(&newProduct)
		if result.Error != nil {
			slog.Error("Failed to create product", "error", result.Error)
			return &pb.CreateProductsResp{
				Result: false,
			}, result.Error
		}
		slog.Info("Product created successfully", "product_id", newProduct.ID)
	}
	return &pb.CreateProductsResp{
		Result: true,
	}, nil
}
