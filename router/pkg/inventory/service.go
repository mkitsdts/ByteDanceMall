package inventory

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/cart"

	"google.golang.org/grpc"
)

var client pb.CartServiceClient

func GetClient() pb.CartServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().Inventory.Host + ":" + strconv.Itoa(config.Get().Inventory.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewCartServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.HasItem(context.Background(), &pb.HasItemReq{
		UserId: 114514,
	})
	if err != nil {
		slog.Error("failed to start cart service", "error: ", err)
		panic(err)
	}
	slog.Info("get cart grpc response", "result: ", resp.HasItem)
}
