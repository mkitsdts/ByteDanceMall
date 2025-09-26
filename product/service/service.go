package service

import (
	pb "bytedancemall/product/proto"
	"log/slog"
)

type ProductService struct {
	pb.UnimplementedProductCatalogServiceServer
}

func NewProductService() (*ProductService, error) {
	s := &ProductService{}
	slog.Info("ProductService initialized successfully")
	return s, nil
}
