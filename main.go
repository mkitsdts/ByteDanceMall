package main

import (
	"bytedancemall/seckill/internal/config"
	"bytedancemall/seckill/internal/grpc"
	"bytedancemall/seckill/internal/service"
	"bytedancemall/seckill/pkg/kafka"
	"bytedancemall/seckill/pkg/redis"
	"bytedancemall/seckill/pkg/tracer"
	"log/slog"
	"net"
)

func Init() (*grpc.Server, error) {
	// load configuration
	config, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return nil, err
	}
	slog.Info("Configuration loaded successfully", "config", config)

	// initialize Jaeger tracer
	tracer, err := tracer.InitJaeger(config)
	if err != nil {
		slog.Error("Failed to initialize Jaeger tracer", "error", err)
		return nil, err
	}
	slog.Info("Jaeger tracer initialized successfully", "tracer", tracer)

	// initialize Redis client
	redisClient, err := redis.NewRedisClusterClient(config)
	if err != nil {
		slog.Error("Failed to initialize Redis client", "error", err)
		return nil, err
	}
	slog.Info("Redis client initialized successfully", "client", redisClient)

	// initialize Kafka producer
	kafkaProducer, err := kafka.InitKafkaProducer(config)
	if err != nil {
		slog.Error("Failed to initialize KafkapProducer", "error", err)
		return nil, err
	}
	slog.Info("Kafka producer initialized successfully", "producer", kafkaProducer)

	// initialize Seckill service
	seckillService, err := service.NewSeckillService(redisClient, kafkaProducer)
	if err != nil {
		slog.Error("Failed to initialize Seckill service", "error", err)
		return nil, err
	}
	slog.Info("Seckill service initialized successfully", "service", seckillService)
	// initialize Grpc server
	grpcServer := grpc.NewGRPCServer(seckillService)
	slog.Info("Grpc server initialized successfully", "server", grpcServer)
	return grpcServer, nil
}

func main() {
	// Initialize the service with the loaded configuration
	grpcServer, err := Init()
	if err != nil {
		slog.Error("Failed to initialize service", "error", err)
		return
	}
	// Start the gRPC server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("Failed to listen on port 50051", "error", err)
		return
	}
	slog.Info("gRPC server is listening on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("Failed to serve gRPC server", "error", err)
		return
	}
	slog.Info("gRPC server stopped gracefully")
	// Graceful shutdown logic can be added here if needed
	slog.Info("Service initialized and gRPC server started successfully")
	defer func() {
		grpcServer.GracefulStop()
		slog.Info("gRPC server stopped gracefully")
	}()
	slog.Info("All resources closed successfully")
}
