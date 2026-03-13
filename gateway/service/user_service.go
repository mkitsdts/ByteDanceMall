package service

import (
	"context"

	userpb "bytedancemall/gateway/proto/user"
	"bytedancemall/gateway/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	UserID   uint64
	Password string
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*userpb.RegisterResp, error) {
	return s.repo.Register(ctx, &userpb.RegisterReq{
		Email:    input.Email,
		Password: input.Password,
	})
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (*userpb.LoginResp, error) {
	return s.repo.Login(ctx, &userpb.LoginReq{
		Email:    input.Email,
		UserId:   input.UserID,
		Password: input.Password,
	})
}
