package main

import (
	"bytedancemall/order/config"
	"bytedancemall/order/pkg/database"
	"bytedancemall/order/pkg/kafka"
	"bytedancemall/order/pkg/redis"
	pb "bytedancemall/order/proto"
	"bytedancemall/order/service"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 加载配置
	if err := config.Init(); err != nil {
		slog.Error("Failed to load config", "error", err)
		time.Sleep(10 * time.Second)
		return
	}

	// 设置监听端口
	port := config.Cfg.Server.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("Failed to listen: %v", "error", err)
		time.Sleep(10 * time.Second)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	// 创建数据库连接
	if err := database.NewDatabase(&service.Order{}); err != nil {
		slog.Error("Failed to connect to database", "error", err)
		time.Sleep(10 * time.Second)
		return
	}

	if err := redis.NewRedis(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		time.Sleep(10 * time.Second)
		return
	}

	if err := kafka.NewKafkaWriter(); err != nil {
		slog.Error("Failed to create Kafka producer", "error", err)
		time.Sleep(10 * time.Second)
		return
	}

	if err := kafka.NewKafkaReader(); err != nil {
		slog.Error("Failed to create Kafka reader", "error", err)
		time.Sleep(10 * time.Second)
		return
	}

	// 创建并注册UserService
	userService := service.NewOrderService()
	pb.RegisterOrderServiceServer(s, userService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Println("用户服务启动成功，监听端口:", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}
