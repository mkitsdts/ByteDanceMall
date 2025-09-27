package order

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/order"

	"google.golang.org/grpc"
)

var client pb.OrderServiceClient

func GetClient() pb.OrderServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().Order.Host + ":" + strconv.Itoa(config.Get().Order.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewOrderServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.ListOrder(context.Background(), &pb.ListOrderReq{
		UserId: 114514,
	})
	if err != nil {
		slog.Error("failed to start order service", "error: ", err)
		panic(err)
	}
	slog.Info("get order grpc response", "result: ", resp.Orders)
}
