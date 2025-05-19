package service

import (
	"bytedancemall/llm/mcp"
	pb "bytedancemall/llm/proto"
	"encoding/json"
	"os"
)

type LLMService struct {
	MCP *mcp.MCPClient
	pb.UnimplementedLLMServiceServer
}

func NewLLMService() *LLMService {
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// 读取配置文件
	decode := json.NewDecoder(file)
	var config Config
	if err := decode.Decode(&config); err != nil {
		panic(err)
	}
	s := &LLMService{}
	s.MCP = mcp.NewMCPClient(config.Model.Name, config.Model.Host, config.Model.Key)
	s.MCP.AddTool("search_products", "search products by keyword from user query",
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
	s.MCP.AddTool("get_product_details", "get product details by product id",
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
	return s
}
