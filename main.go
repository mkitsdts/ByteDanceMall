package main

import (
	"bytedancemall/auth/config"
	"bytedancemall/auth/pkg"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/service"
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 设置监听端口
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to load config: %v", "error: ", err)
		return
	}
	// 初始化数据库连接

	db, err := pkg.NewDatabase(cfg)
	if err != nil {
		slog.Error("Failed to connect to database: %v", "error: ", err)
		return
	}
	// 初始化Redis连接
	redisCluster := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    cfg.Redis.Host,
		Password: cfg.Redis.Password,
	})
	if err := redisCluster.Ping(context.Background()).Err(); err != nil {
		slog.Error("Failed to connect to Redis: %v", err)
		return
	}
	// 创建并注册UserService
	authService := service.NewAuthService(redisCluster, db)
	pb.RegisterAuthServiceServer(s, authService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Printf("用户服务启动成功，监听端口: %d", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v", err)
	}
}
