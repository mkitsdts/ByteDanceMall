package service

import (
	"bytedancemall/user/config"
	pb "bytedancemall/user/proto"
	"fmt"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
}

func NewUserService() (*UserService, error) {

	if err := config.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize configs: %w", err)
	}

	return &UserService{}, nil
}
