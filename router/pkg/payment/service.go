package payment

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/payment"

	"google.golang.org/grpc"
)

var client pb.PaymentServiceClient

func GetClient() pb.PaymentServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().Payment.Host + ":" + strconv.Itoa(config.Get().Payment.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewPaymentServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.QueryStatus(context.Background(), &pb.QueryStatusReq{
		OrderId: 114514,
	})
	if err != nil {
		slog.Error("failed to start payment service", "error: ", err)
		panic(err)
	}
	slog.Info("get payment grpc response", "result: ", resp.Status)
}
