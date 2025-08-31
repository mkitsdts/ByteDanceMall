package main

import (
	"bytedancemall/inventory/config"
	"bytedancemall/inventory/pkg"
	pb "bytedancemall/inventory/proto"
	"bytedancemall/inventory/service"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := 14802

	// 设置监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Failed to read config: %v", err)
		return
	}

	database, err := pkg.NewDatabase(&cfg.Database)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v", err)
		return
	}

	// redis, err := pkg.NewRedisClusterClient(&cfg.Redis)
	// if err != nil {
	// 	fmt.Printf("Failed to initialize redis: %v", err)
	// 	return
	// }

	redis, err := pkg.NewRedisClient(&cfg.Redis)
	if err != nil {
		fmt.Printf("Failed to initialize redis: %v", err)
		return
	}

	reader, err := pkg.NewKafkaReader(&cfg.KafkaReader)
	if err != nil {
		fmt.Printf("Failed to initialize kafka: %v", err)
		return
	}

	writer, err := pkg.NewKafkaWriter(&cfg.KafkaWriter)
	if err != nil {
		fmt.Printf("Failed to initialize kafka writer: %v", err)
		return
	}

	// 创建并注册InventoryService
	service := service.NewInventoryService(database, redis, writer, reader)
	pb.RegisterInventoryServiceServer(s, service)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Println("用户服务启动成功，监听端口:", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v", err)
	}
}
