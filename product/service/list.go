package service

import (
	"bytedancemall/product/model"
	"bytedancemall/product/pkg/database"
	pb "bytedancemall/product/proto"
	"context"
	"fmt"
	"log/slog"
	"time"
)

// 获取单个产品
func (s *ProductService) GetProductDetail(ctx context.Context, req *pb.GetProductDetailReq) (*pb.GetProductDetailResp, error) {
	var product model.Product
	maxRetries := 3
	for i := range maxRetries {
		result := database.Get().First(&product, req.ProductId)
		if result.Error == nil {
			slog.Info("Fetched product successfully")
			break
		}
		duration := i << 1
		if i == 2 {
			slog.Error("Failed to fetch product after retries", "error", result.Error)
			return nil, fmt.Errorf("failed to fetch product: %w", result.Error)
		}
		time.Sleep(time.Duration(duration) * time.Second)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	go s.asyncEnsureRedisSet(ctx, fmt.Sprint(product.ID), product)
	return &pb.GetProductDetailResp{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
	}, nil
}

// 根据页数获取商品
func (s *ProductService) ListProducts(ctx context.Context, req *pb.ListProductsReq) (*pb.ListProductsResp, error) {
	// 计算偏移量
	offset := ((int64(req.Page)) - 1) * req.PageSize
	// 从数据库中获取商品
	var products []model.Product

	maxRetries := 3
	for i := range maxRetries {
		result := database.Get().Limit(int(req.PageSize)).Offset(int(offset)).Find(&products)
		if result.Error == nil {
			slog.Info("Fetched products successfully", "count", len(products))
			break
		}
		duration := i << 1
		if i == 2 {
			slog.Error("Failed to fetch products after retries", "error", result.Error)
			return nil, fmt.Errorf("failed to fetch products: %w", result.Error)
		}
		time.Sleep(time.Duration(duration) * time.Second)
	}
	var respProducts []*pb.Product
	for _, v := range products {
		respProducts = append(respProducts, &pb.Product{
			Id:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			Price:       v.Price,
		})
	}
	// 返回商品
	return &pb.ListProductsResp{
		Products: respProducts,
	}, nil
}
