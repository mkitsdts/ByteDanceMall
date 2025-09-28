package main

import (
	"bytedancemall/payment/config"
	"bytedancemall/payment/pkg/database"
	"bytedancemall/payment/pkg/kafka"
	"bytedancemall/payment/pkg/redis"
	pb "bytedancemall/payment/proto"
	"bytedancemall/payment/service"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 设置监听端口
	port := config.Cfg.Server.Port

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Failed to listen:", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()

	// 初始化数据库连接
	if err := database.NewDatabase(); err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}

	// 初始化Kafka Reader
	if err := kafka.NewKafkaReader(); err != nil {
		fmt.Println("Failed to connect to Kafka reader:", err)
		return
	}

	// 初始化Kafka Writer
	if err := kafka.NewKafkaWriter(); err != nil {
		fmt.Println("Failed to connect to Kafka writer:", err)
		return
	}
	// 初始化Redis集群
	if err := redis.NewRedis(); err != nil {
		fmt.Printf("Failed to connect to Redis cluster: %v", err)
		return
	}

	// RegisterPaymentServiceServer
	userService := service.NewPaymentService()
	pb.RegisterPaymentServiceServer(s, userService)

	// 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(s)

	fmt.Println("用户服务启动成功，监听端口:", port)

	// 启动服务
	if err := s.Serve(lis); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}
