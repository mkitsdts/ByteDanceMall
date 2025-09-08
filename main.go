package main

import (
	"bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/service"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	port := 14801

	redis.InitRedis()

	// 设置监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	// 创建并注册UserService
	authService := service.NewAuthService()
	pb.RegisterAuthServiceServer(s, authService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Printf("用户服务启动成功，监听端口: %d", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v", err)
	}
}
