package main

import (
	"bytedancemall/user/config"
	"bytedancemall/user/model"
	dbpkg "bytedancemall/user/pkg/database"
	redispkg "bytedancemall/user/pkg/redis"
	pb "bytedancemall/user/proto"
	"bytedancemall/user/repository"
	"bytedancemall/user/service"
	"bytedancemall/user/usecase"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := config.Init(); err != nil {
		slog.Error("failed to initialize config", "error", err)
		return
	}

	database, err := dbpkg.New(&model.User{}, &model.LoginRecord{})
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		return
	}

	redisClient, err := redispkg.NewClient()
	if err != nil {
		slog.Error("failed to initialize redis", "error", err)
		return
	}

	repos := repository.New(database.Master, redisClient)
	userUsecase := usecase.New(repos)
	userService := service.New(userUsecase)

	port := config.Conf.Server.Port
	if port == 0 {
		port = 50051
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("failed to listen", "port", port, "error", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userService)
	reflection.Register(s)
	slog.Info("user service started", "port", port)

	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
	}
}
