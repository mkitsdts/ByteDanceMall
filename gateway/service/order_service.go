package service

import (
	"context"

	orderpb "bytedancemall/gateway/proto/order"
	"bytedancemall/gateway/repository"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

type CreateOrderInput struct {
	UserID    uint64
	ProductID uint64
	Amount    uint64
	OrderID   uint64
	Cost      float32
	Street    string
	City      string
	State     string
}

func (s *OrderService) Create(ctx context.Context, input CreateOrderInput) (*orderpb.CreateOrderResp, error) {
	return s.repo.CreateOrder(ctx, &orderpb.CreateOrderReq{
		UserId:    input.UserID,
		ProductId: input.ProductID,
		Amount:    input.Amount,
		OrderId:   input.OrderID,
		Cost:      input.Cost,
		Address: &orderpb.Address{
			StreetAddress: input.Street,
			City:          input.City,
			State:         input.State,
		},
	})
}
