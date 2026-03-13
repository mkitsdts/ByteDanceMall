package repository

import (
	"context"

	"bytedancemall/gateway/config"
	orderpb "bytedancemall/gateway/proto/order"
)

type OrderRepository struct {
	client *grpcClient
}

func NewOrderRepository(cfg config.OrderService) *OrderRepository {
	return &OrderRepository{client: newGRPCClient(cfg.Host, cfg.Port)}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, req *orderpb.CreateOrderReq) (*orderpb.CreateOrderResp, error) {
	conn, err := r.client.Conn()
	if err != nil {
		return nil, err
	}
	return orderpb.NewOrderServiceClient(conn).CreateOrder(ctx, req)
}
