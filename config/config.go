package config

import (
	"encoding/json"
	"os"
)

type DatabaseConfig struct {
	Master       string   `json:"master"`
	Slaves       []string `json:"slaves"`
	Name         string   `json:"name"`
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	MaxIdleConns int      `json:"max_idle_conns"`
	MaxOpenConns int      `json:"max_open_conns"`
}

type RedisConfig struct {
	Host     []string `json:"host"`
	Port     int      `json:"port"`
	Password string   `json:"password"`
}

type KafkaWriter struct {
	Host     []string `json:"host"`
	Port     string   `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Topic    []string `json:"topic"`
	GroupID  string   `json:"group_id"`
}

type KafkaReader struct {
	Host     []string `json:"host"`
	Port     string   `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Topic    []string `json:"topic"`
	GroupID  string   `json:"group_id"`
}

type Server struct {
	Port int `json:"port"`
}

type Config struct {
	Server      Server         `json:"server"`
	Database    DatabaseConfig `json:"database"`
	Redis       RedisConfig    `json:"redis"`
	KafkaWriter KafkaWriter    `json:"kafka_writer"`
	KafkaReader KafkaReader    `json:"kafka_reader"`
}

func ReadConfig() (*Config, error) {
	var config Config
	file, err := os.Open("configs.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
