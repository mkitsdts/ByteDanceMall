package service

import (
	"bytedancemall/llm/cllm"
	"encoding/json"
	"os"
)

func NewLLMService() *LLMService {
	type ModelConfig struct {
		Host string `json:"host"`
		Name string `json:"name"`
	}
	type Config struct {
		Model ModelConfig `json:"model"`
	}
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	config := Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}
	model := cllm.ModelService{}
	model.Init(config.Model.Host, config.Model.Name)
	return &LLMService{
		Model: &model,
	}
}
