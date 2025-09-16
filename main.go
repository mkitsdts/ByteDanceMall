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
	// 初始化
	config.Init()
	rds.Init()

	server := grpc.NewServer()

	s := service.NewLLMService()
	pb.RegisterLLMServiceServer(server, s)

	// // 注册reflection服务，便于使用grpcurl等工具调试
	reflection.Register(server)

	// // 监听端口
	port := 14804
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	fmt.Println("Listening on port:", port)
	if err != nil {
		fmt.Println("监听端口失败:", err)
		return
	}
	//
	// // 启动服务
	//
	if err := server.Serve(listener); err != nil {
		fmt.Println("启动服务失败: ", err)
		return
	}
	//
	fmt.Println("用户服务启动成功，监听端口:", port)
}
