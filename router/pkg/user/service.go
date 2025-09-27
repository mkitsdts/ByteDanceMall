package user

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/user"

	"google.golang.org/grpc"
)

var client pb.UserServiceClient

func GetClient() pb.UserServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().User.Host + ":" + strconv.Itoa(config.Get().User.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewUserServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.GetUserInfo(context.Background(), &pb.GetUserInfoReq{
		UserId: 114514,
	})
	if err != nil {
		slog.Error("failed to start user service", "error: ", err)
		panic(err)
	}
	slog.Info("get user grpc response", "result: ", resp.Exists)
}
