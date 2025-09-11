package main

import (
	"bytedancemall/llm/config"
	rds "bytedancemall/llm/pkg/redis"
	pb "bytedancemall/llm/proto"
	"bytedancemall/llm/service"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize the gRPC server
	server := grpc.NewServer()

	// 	// 注册服务

	config.Init()
	rds.Init()

	llmserever := service.NewLLMService()

	pb.RegisterLLMServiceServer(server, llmserever)

	// // 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(server)
	// // 监听端口
	port := 14804
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	//
	if err != nil {
		fmt.Printf("监听端口失败: %v", err)
		return
	}
	//
	// // 启动服务
	//
	if err := server.Serve(listener); err != nil {
		fmt.Printf("启动服务失败: %v", err)
		return
	}
	//
	// fmt.Printf("用户服务启动成功，监听端口: %d", port)
}
