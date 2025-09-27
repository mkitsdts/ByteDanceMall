package product

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/product"

	"google.golang.org/grpc"
)

var client pb.ProductCatalogServiceClient

func GetClient() pb.ProductCatalogServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().Product.Host + ":" + strconv.Itoa(config.Get().Product.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewProductCatalogServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.GetProductDetail(context.Background(), &pb.GetProductDetailReq{
		ProductId: 114514,
	})
	if err != nil {
		slog.Error("failed to start product catalog service", "error: ", err)
		panic(err)
	}
	slog.Info("get product catalog grpc response", "result: ", resp.Product)
}
