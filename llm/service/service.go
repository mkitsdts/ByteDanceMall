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
	ChatModel  *mcp.ChatModel
	EmbedModel *mcp.ChatModel
	pb.UnimplementedLLMServiceServer
}

func (s *LLMService) InitLLMModel() {
	s.ChatModel = mcp.NewChatModel()
	s.ChatModel.SetChatModel(config.Conf.LLM.Name, config.Conf.LLM.Host, config.Conf.LLM.Key)
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

func (s *LLMService) InitEmbedModel() {
	s.EmbedModel = mcp.NewChatModel()
	s.EmbedModel.SetEmbedModel(config.Conf.Embed.Name, config.Conf.Embed.Host, config.Conf.Embed.Key)
}

func NewLLMService() *LLMService {
	s := &LLMService{}
	s.InitLLMModel()
	s.InitEmbedModel()
	path := "doc"
	index := 0
	if err := rds.CreateIndex(); err != nil {
		panic(err)
	}
	fmt.Println("Document ingestion completed.")
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			fmt.Println("Processing file:", path)
			// 根据一级标签分片
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(content)
			re := regexp.MustCompile(`(?m)^# .*$`)
			headerIndices := re.FindAllStringIndex(text, -1)
			text = strings.ReplaceAll(text, "\n", " ")
			// 没有一级标题，整体作为一个片段
			if len(headerIndices) == 0 {
				vec, err := s.EmbedModel.Embedding(context.Background(), text)
				if err != nil {
					return err
				}
				fmt.Println("preface embedding length:", len(*vec))
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

			// 跳过前言部分
			if headerIndices[0][0] > 0 {
				preface := text[:headerIndices[0][0]]
				fmt.Println("Preface found:", preface)
				if strings.TrimSpace(preface) == "" {
					return nil
				}
				vec, err := s.EmbedModel.Embedding(context.Background(), preface)
				if err != nil {
					return err
				}
				fmt.Println("preface embedding length:", len(*vec))
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
				headerText := utils.RewriteSentence(text[start:indices[1]])
				fmt.Println("Header found:", headerText)
				fmt.Println("Section content:", section)
				vec, err := s.EmbedModel.Embedding(context.Background(), section)
				if err != nil {
					return err
				}

				fmt.Println("section embedding length:", len(*vec))
				buffer := utils.FloatsToBytes(vec)
				_, err = rds.GetRedisClient().HSet(context.Background(),
					fmt.Sprintf("doc:%v", index),
					map[string]any{
						"content":   headerText + "\n" + section,
						"embedding": buffer,
					},
				).Result()

				fmt.Printf("DEBUG_INGESTION: Key: doc:%d, Vector Bytes (first 16): %x\n", index, buffer[:16])

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
