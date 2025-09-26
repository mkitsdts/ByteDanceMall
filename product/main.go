package main

import (
	"bytedancemall/product/config"
	"bytedancemall/product/pkg/database"
	"bytedancemall/product/pkg/redis"
	pb "bytedancemall/product/proto"
	"bytedancemall/product/service"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// 初始化配置
	if err := config.Init(); err != nil {
		slog.Error("Failed to initialize config: %v", "error", err)
		return
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		slog.Error("Failed to initialize database: %v", "error", err)
		return
	}

	// 初始化Redis
	if err := redis.NewRedisCluster(); err != nil {
		slog.Error("Failed to initialize redis: %v", "error", err)
		return
	}

	s := grpc.NewServer()
	productService, err := service.NewProductService()
	if err != nil {
		slog.Error("Failed to create product service: %v", "error", err)
		return
	}
	pb.RegisterProductCatalogServiceServer(s, productService)

	// 设置监听端口
	port := 14801
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
	}

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	slog.Info("grpc service start successful", "port", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		slog.Error("Failed to serve: %v", "error", err)
	}
}
