package main

import (
	pb "bytedancemall/product/proto"
	"bytedancemall/product/service"
	"fmt"
	"net"

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

	// 创建并注册UserService
	userService := service.NewProductService()
	pb.RegisterProductCatalogServiceServer(s, userService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Printf("用户服务启动成功，监听端口: %d", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v", err)
	}
}