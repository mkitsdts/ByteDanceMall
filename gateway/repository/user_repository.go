package repository

import (
	"context"

	"bytedancemall/gateway/config"
	userpb "bytedancemall/gateway/proto/user"
)

type UserRepository struct {
	client *grpcClient
}

func NewUserRepository(cfg config.UserService) *UserRepository {
	return &UserRepository{client: newGRPCClient(cfg.Host, cfg.Port)}
}

func (r *UserRepository) Register(ctx context.Context, req *userpb.RegisterReq) (*userpb.RegisterResp, error) {
	conn, err := r.client.Conn()
	if err != nil {
		return nil, err
	}
	return userpb.NewUserServiceClient(conn).Register(ctx, req)
}

func (r *UserRepository) Login(ctx context.Context, req *userpb.LoginReq) (*userpb.LoginResp, error) {
	conn, err := r.client.Conn()
	if err != nil {
		return nil, err
	}
	return userpb.NewUserServiceClient(conn).Login(ctx, req)
}
