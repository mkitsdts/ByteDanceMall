package service

import (
	"bytedancemall/llm/config"
	rds "bytedancemall/llm/pkg/redis"
	pb "bytedancemall/llm/proto"
	"bytedancemall/llm/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mcp "github.com/mkitsdts/mcp-go/chatmodel"
)

type LLMService struct {
	ChatModel *mcp.ChatModel
	pb.UnimplementedLLMServiceServer
}

func (s *LLMService) InitLLMModel() {
	s.ChatModel = mcp.NewChatModel()
	s.ChatModel.SetChatModel(config.Conf.LLM.Name, config.Conf.LLM.Host, config.Conf.LLM.Key)
	s.ChatModel.SetEmbedModel(config.Conf.Embed.Name, config.Conf.Embed.Host, config.Conf.Embed.Key)
	s.ChatModel.AddTool("search_products", "search products by keyword from user query",
		mcp.Paramaters{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The user input to search for products",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "The page number to return",
				},
			},
			Required: []string{"query"},
		},
		SearchProducts,
	)
	s.ChatModel.AddTool("get_product_details", "get product details by product id",
		mcp.Paramaters{
			Type: "object",
			Properties: map[string]any{
				"product_id": map[string]any{
					"type":        "string",
					"description": "The product id to get details for",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "The page number to return",
				},
			},
			Required: []string{"product_id"},
		},
		GetProductDetails,
	)
}

func (s *LLMService) InitAssistant() {
	s.ChatModel.SetEmbedModel(config.Conf.Embed.Name, config.Conf.Embed.Host, config.Conf.Embed.Key)

}

func NewLLMService() *LLMService {
	s := &LLMService{}
	s.InitLLMModel()
	s.InitAssistant()
	path := "doc"
	index := 0
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			// 根据一级标签分片
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(content)
			re := regexp.MustCompile(`(?m)^# .*$`)
			headerIndices := re.FindAllStringIndex(text, -1)

			if len(headerIndices) == 0 {
				vec, err := s.ChatModel.Embedding(context.Background(), text)
				if err != nil {
					return err
				}
				buffer := utils.FloatsToBytes(vec)
				_, err = rds.GetRedisClient().HSet(context.Background(),
					fmt.Sprintf("doc:%v", index),
					map[string]any{
						"content":   text,
						"embedding": buffer,
					},
				).Result()
				index++
				if err != nil {
					return err
				}
				return nil
			}

			// Handle content before the first header if there is any
			if headerIndices[0][0] > 0 {
				preface := text[:headerIndices[0][0]]
				if strings.TrimSpace(preface) == "" {
					return nil
				}
				vec, err := s.ChatModel.Embedding(context.Background(), preface)
				if err != nil {
					return err
				}
				buffer := utils.FloatsToBytes(vec)
				_, err = rds.GetRedisClient().HSet(context.Background(),
					fmt.Sprintf("doc:%v", index),
					map[string]any{
						"content":   preface,
						"embedding": buffer,
					},
				).Result()
				index++
				if err != nil {
					return err
				}
			}

			// Process each section defined by headers
			for i, indices := range headerIndices {
				start := indices[0]
				end := len(text)
				if i < len(headerIndices)-1 {
					end = headerIndices[i+1][0]
				}

				section := text[start:end]
				headerText := strings.TrimSpace(text[indices[0]:indices[1]])
				if strings.TrimSpace(section) == "" {
					continue
				}
				vec, err := s.ChatModel.Embedding(context.Background(), section)
				if err != nil {
					return err
				}
				buffer := utils.FloatsToBytes(vec)
				_, err = rds.GetRedisClient().HSet(context.Background(),
					fmt.Sprintf("doc:%v", index),
					map[string]any{
						"content":   headerText + "\n" + section,
						"embedding": buffer,
					},
				).Result()
				index++
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return s
}
