package main

import (
	"bytedancemall/order/config"
	"bytedancemall/order/pkg"
	pb "bytedancemall/order/proto"
	"bytedancemall/order/service"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 设置监听端口
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("Failed to listen: %v", "error", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return
	}
	// 创建数据库连接
	db, err := pkg.NewDatabase(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return
	}

	redis, err := pkg.NewRedis(cfg)
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		return
	}

	kafkaWriter, err := pkg.NewKafkaProducer(cfg)
	if err != nil {
		slog.Error("Failed to create Kafka producer", "error", err)
		return
	}

	kafkaReader, err := pkg.NewKafkaReader(cfg)
	if err != nil {
		slog.Error("Failed to create Kafka reader", "error", err)
		return
	}

	// 创建并注册UserService
	userService := service.NewOrderService(db, redis, kafkaWriter, kafkaReader)
	pb.RegisterOrderServiceServer(s, userService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Printf("用户服务启动成功，监听端口: %d", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v", err)
	}
}
