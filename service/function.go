package service

import (
	pb "bytedancemall/llm/proto"
	"bytedancemall/llm/utils"
	"context"
)

func (s *LLMService) FilterKeyWords(ctx context.Context, req *pb.FilterKeyWordReq) (*pb.FilterKeyWordResp, error) {
	keyword, err := s.Model.Chat(req.Keyword)
	if err != nil {
		return nil, err
	}
	return &pb.FilterKeyWordResp{
		Result: keyword,
	}, nil
}

func (s *LLMService) ChooseProduct(ctx context.Context, req *pb.ChooseProductReq) (*pb.ChooseProductResp, error) {
	res, err := s.Model.Chat(utils.ProductInfoToString(req.ProductInfo))
	if err != nil {
		return nil, err
	}
	id := utils.StringParseToUID(res)
	if len(id) == 0 {
		return nil, nil
	}
	return &pb.ChooseProductResp{
		ProductId: id,
	}, nil
}
