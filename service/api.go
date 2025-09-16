package service

import (
	rds "bytedancemall/llm/pkg/redis"
	pb "bytedancemall/llm/proto"
	"bytedancemall/llm/utils"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (s *LLMService) ShopAssistant(ctx context.Context, req *pb.QuestionReq) (*pb.AssistantAnswerResp, error) {
	answer, err := s.ChatModel.Chat(ctx, req.Question)
	if err != nil {
		return &pb.AssistantAnswerResp{
			Result: false,
		}, err
	}
	return &pb.AssistantAnswerResp{
		Result: true,
		Answer: answer,
	}, nil
}

func (s *LLMService) IntelligentCustomer(ctx context.Context, req *pb.QuestionReq) (*pb.CustomerAnswerResp, error) {
	// get embedding from question
	rewrittenQuestion := utils.RewriteSentence(req.Question)
	embedding, err := s.EmbedModel.Embedding(context.Background(), rewrittenQuestion)
	if err != nil {
		return nil, err
	}
	fmt.Println("User question:", rewrittenQuestion)
	// for _, v := range *embedding {
	// 	fmt.Print(v, " ")
	// }
	fmt.Println()

	data := utils.FloatsToBytes(embedding)

	fmt.Printf("DEBUG_QUERY: Vector Bytes (first 16): %x\n", data[:16])
	results, err := rds.GetRedisClient().FTSearchWithArgs(ctx,
		"smart_assistant_idx",
		"*=>[KNN 2 @embedding $vec AS vector_distance]",
		&redis.FTSearchOptions{
			Return: []redis.FTSearchReturn{
				{FieldName: "vector_distance"},
				{FieldName: "content"},
			},
			DialectVersion: 2,
			Params: map[string]any{
				"vec": data,
			},
		},
	).Result()

	if err != nil {
		return &pb.CustomerAnswerResp{
			Result: false,
		}, err
	}

	context := `I found some related information for you.
	You can use the following information to answer the question.
	And you can't use any other information and tools.
	Just answer and don't explain anything.
	`

	if len(results.Docs) == 0 {
		fmt.Println("No related documents found.")
		return &pb.CustomerAnswerResp{
			Result: false,
		}, nil
	} else {
		fmt.Println("Found related documents:")
		for _, doc := range results.Docs {
			fmt.Printf(
				"ID: %v, Distance:%v, Content:'%v'\n",
				doc.ID, doc.Fields["vector_distance"], doc.Fields["content"],
			)
		}
	}

	for _, doc := range results.Docs {
		context += fmt.Sprintf(
			"ID: %v, Distance:%v, Content:'%v'\n",
			doc.ID, doc.Fields["vector_distance"], doc.Fields["content"],
		)
	}
	answer, err := s.ChatModel.Chat(ctx, context)
	if err != nil {
		return &pb.CustomerAnswerResp{
			Result: false,
		}, err
	}
	return &pb.CustomerAnswerResp{
		Result: true,
		Answer: answer,
	}, nil
}
