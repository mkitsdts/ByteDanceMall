package service

type ModelConfig struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Key  string `json:"key"`
}

type Config struct {
	Model ModelConfig `json:"model"`
}
