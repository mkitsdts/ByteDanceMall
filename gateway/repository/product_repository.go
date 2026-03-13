package repository

import (
	"context"

	"bytedancemall/gateway/config"
	productpb "bytedancemall/gateway/proto/product"
)

type ProductRepository struct {
	client *grpcClient
}

func NewProductRepository(cfg config.ProductService) *ProductRepository {
	return &ProductRepository{client: newGRPCClient(cfg.Host, cfg.Port)}
}

func (r *ProductRepository) ListProducts(ctx context.Context, req *productpb.ListProductsReq) (*productpb.ListProductsResp, error) {
	conn, err := r.client.Conn()
	if err != nil {
		return nil, err
	}
	return productpb.NewProductCatalogServiceClient(conn).ListProducts(ctx, req)
}

func (r *ProductRepository) GetProductDetail(ctx context.Context, req *productpb.GetProductDetailReq) (*productpb.GetProductDetailResp, error) {
	conn, err := r.client.Conn()
	if err != nil {
		return nil, err
	}
	return productpb.NewProductCatalogServiceClient(conn).GetProductDetail(ctx, req)
}
