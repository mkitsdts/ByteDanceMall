package service

import (
	"context"

	productpb "bytedancemall/gateway/proto/product"
	"bytedancemall/gateway/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

type ListProductsInput struct {
	Page         int32
	PageSize     int64
	CategoryName string
}

func (s *ProductService) List(ctx context.Context, input ListProductsInput) (*productpb.ListProductsResp, error) {
	return s.repo.ListProducts(ctx, &productpb.ListProductsReq{
		Page:         input.Page,
		PageSize:     input.PageSize,
		CategoryName: input.CategoryName,
	})
}

func (s *ProductService) GetDetail(ctx context.Context, productID uint64) (*productpb.GetProductDetailResp, error) {
	return s.repo.GetProductDetail(ctx, &productpb.GetProductDetailReq{ProductId: productID})
}
