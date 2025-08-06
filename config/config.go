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
	Port     string   `json:"port"`
	Password string   `json:"password"`
}

type KafkaWriterConfig struct {
	Host     []string `json:"host"`
	Port     string   `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Topic    string   `json:"topic"`
}

type KafkaReaderConfig struct {
	Host     []string `json:"host"`
	Port     string   `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Topic    string   `json:"topic"`
}

type Config struct {
	Database    DatabaseConfig    `json:"database"`
	Redis       RedisConfig       `json:"redis"`
	KafkaWriter KafkaWriterConfig `json:"kafka_writer"`
	KafkaReader KafkaReaderConfig `json:"kafka_reader"`
}

func ReadConfig() (*Config, error) {
	var config Config
	file, err := os.Open("config.json")
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
