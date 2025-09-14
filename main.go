package main

import (
	"bytedancemall/auth/config"
	"bytedancemall/auth/model"
	"bytedancemall/auth/pkg/database"
	rds "bytedancemall/auth/pkg/redis"
	pb "bytedancemall/auth/proto"
	"bytedancemall/auth/service"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// 加载配置
	if err := config.Init(); err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	port := config.Conf.Server.Port

	for i := range 100 {
		// 重试连接，防止 redis 未启动导致连接失败
		if err := rds.InitRedis(); err == nil {
			break
		} else if i == 99 {
			fmt.Println("Failed to initialize Redis:", err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	for i := range 100 {
		// 重试连接，防止 mysql 未启动导致连接失败
		if err := database.NewDatabase(&model.RefreshToken{}); err == nil {
			break
		} else if i == 99 {
			fmt.Println("Failed to initialize database:", err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
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
