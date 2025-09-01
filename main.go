package main

import (
	"bytedancemall/cart/config"
	"bytedancemall/cart/pkg"
	pb "bytedancemall/cart/proto"
	"bytedancemall/cart/service"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 设置监听端口
	port := 14804
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Failed to listen: %v", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	config.Init()

	pkg.NewRedis()
	pkg.NewDatabase()

	// 创建并注册UserService
	userService, err := service.NewCartService()
	if err != nil {
		fmt.Println("Failed to create user service: %v", err)
		return
	}
	pb.RegisterCartServiceServer(s, userService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Println("用户服务启动成功，监听端口:", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}
