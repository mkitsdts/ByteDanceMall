package service

import (
	"bytedancemall/llm/cllm"
	pb "bytedancemall/llm/proto"
)

type LLMService struct {
	Model *cllm.ModelService
	pb.UnimplementedLLMServiceServer
}
