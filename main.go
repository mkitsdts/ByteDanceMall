package main

import (
	rds "bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/service"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	port := 14801

	rds.InitRedis()

	// 设置监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Failed to listen:", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	// 创建并注册UserService
	authService := service.NewAuthService()
	pb.RegisterAuthServiceServer(s, authService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Println("用户服务启动成功，监听端口:", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}
