package auth

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/auth"

	"google.golang.org/grpc"
)

var client pb.AuthServiceClient

func GetClient() pb.AuthServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().Auth.Host + ":" + strconv.Itoa(config.Get().Auth.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewAuthServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.VerifyToken(context.Background(), &pb.VerifyTokenReq{
		Token: "test",
	})
	if err != nil {
		slog.Error("failed to start auth service", "error: ", err)
		panic(err)
	}
	slog.Info("get auth grpc response", "result: ", resp.Result)
}
