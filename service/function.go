package service

import (
	pb "bytedancemall/llm/proto"
	"context"
)

func (s *LLMService) FilterKeyWords(ctx context.Context, req *pb.FilterKeyWordReq) (*pb.FilterKeyWordResp, error) {
	s.MCP.Chat(req.Keyword)
	return nil, nil
}

func (s *LLMService) ChooseProduct(ctx context.Context, req *pb.ChooseProductReq) (*pb.ChooseProductResp, error) {
	return nil, nil
}
