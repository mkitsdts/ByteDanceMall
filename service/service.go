package service

import (
	"bytedancemall/user/config"
	"bytedancemall/user/pkg"
	pb "bytedancemall/user/proto"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	Db    *pkg.Database
	Redis *redis.ClusterClient
	pb.UnimplementedUserServiceServer
}

func NewUserService() (*UserService, error) {
	cfg, err := config.InitConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configs: %w", err)
	}

	db, err := pkg.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	redisClient, err := pkg.NewRedisClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &UserService{
		Db:    db,
		Redis: redisClient,
	}, nil
}
