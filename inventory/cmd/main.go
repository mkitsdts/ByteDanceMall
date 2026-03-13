package main

import (
	"bytedancemall/inventory/cache"
	"bytedancemall/inventory/config"
	"bytedancemall/inventory/model"
	dbpkg "bytedancemall/inventory/pkg/database"
	kafkapkg "bytedancemall/inventory/pkg/kafka"
	redispkg "bytedancemall/inventory/pkg/redis"
	pb "bytedancemall/inventory/proto"
	"bytedancemall/inventory/repository"
	"bytedancemall/inventory/service"
	"bytedancemall/inventory/usecase"
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

	if err := config.Init(); err != nil {
		fmt.Printf("Failed to read config: %v", err)
		return
	}

	database, err := dbpkg.New(&model.Inventory{}, &model.OutInventory{})
	if err != nil {
		fmt.Printf("Failed to initialize database: %v", err)
		return
	}

	// redis, err := redispkg.NewClusterClient()
	// if err != nil {
	// 	fmt.Printf("Failed to initialize redis: %v", err)
	// 	return
	// }

	redis, err := redispkg.NewClient()
	if err != nil {
		fmt.Printf("Failed to initialize redis: %v", err)
		return
	}

	reader, err := kafkapkg.NewReader()
	if err != nil {
		fmt.Printf("Failed to initialize kafka: %v", err)
		return
	}

	repos := repository.New(database.Master)
	cacheStore := cache.New(redis)
	inventoryUsecase := usecase.New(repos, cacheStore)

	// 创建并注册InventoryService
	inventoryService := service.NewInventoryService(inventoryUsecase, reader)
	pb.RegisterInventoryServiceServer(s, inventoryService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Println("用户服务启动成功，监听端口:", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v", err)
	}
}
