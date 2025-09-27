package llm

import (
	"context"
	"log/slog"
	"strconv"

	"bytedancemall/router/config"
	pb "bytedancemall/router/proto/llm"

	"google.golang.org/grpc"
)

var client pb.LLMServiceClient

func GetClient() pb.LLMServiceClient {
	return client
}

func init() {
	// 1.创建连接器
	host := config.Get().LLM.Host + ":" + strconv.Itoa(config.Get().LLM.Port)
	conn, err := grpc.NewClient(host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 2.创建对应 grpc 客户端
	client = pb.NewLLMServiceClient(conn)
	// 3.根据函数调用发送 grpc 请求
	resp, err := client.ShopAssistant(context.Background(), &pb.QuestionReq{
		Question: "no reply",
	})
	if err != nil {
		slog.Error("failed to start llm service", "error: ", err)
		panic(err)
	}
	slog.Info("get llm grpc response", "result: ", resp.Answer)
}
