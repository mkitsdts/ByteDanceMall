package grpc

import (
	"bytedancemall/seckill/internal/interceptor"
	"bytedancemall/seckill/internal/service"
	pb "bytedancemall/seckill/proto"

	"google.golang.org/grpc"
	reflction "google.golang.org/grpc/reflection"
)

// Server 是gRPC服务器的包装
type Server struct {
	*grpc.Server
}

// NewGRPCServer 创建并配置gRPC服务器
func NewGRPCServer(seckilllSvc *service.SeckillSer) *Server {
	// 创建拦截器
	tracerInterceptor := interceptor.NewTracerInterceptor()

	// 创建gRPC服务器，注册所有拦截器
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			tracerInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(
			tracerInterceptor.Stream(),
		),
	)

	// 注册服务
	pb.RegisterSeckillServiceServer(grpcServer, seckilllSvc)

	reflction.Register(grpcServer)

	return &Server{Server: grpcServer}
}
