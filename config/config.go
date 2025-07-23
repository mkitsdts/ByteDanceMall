package config

import (
	"encoding/json"
	"log/slog"
	"os"
)

type MysqlConfig struct {
	Host     []string `json:"host"`
	Port     string   `json:"port"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	Database string   `json:"database"`
}

type RedisConfig struct {
	Host     []string `json:"configs"`
	Port     int      `json:"port"`
	Password string   `json:"password"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

type Config struct {
	MysqlConfig MysqlConfig `json:"mysql"`
	RedisConfig RedisConfig `json:"redis"`
	Kafka       KafkaConfig `json:"kafka"`
}

func NewConfig() (*Config, error) {
	file, err := os.ReadFile("config.json")
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		slog.Error("Failed to unmarshal config", "error", err)
		return nil, err
	}
	return &cfg, nil
}
