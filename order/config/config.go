package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var Cfg *Config = &Config{}

type DatabaseConfig struct {
	Master       string   `json:"master" yaml:"master"`
	Slaves       []string `json:"slaves" yaml:"slaves"`
	Name         string   `json:"name" yaml:"name"`
	Port         int      `json:"port" yaml:"port"`
	Username     string   `json:"username" yaml:"username"`
	Password     string   `json:"password" yaml:"password"`
	MaxIdleConns int      `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int      `json:"max_open_conns" yaml:"max_open_conns"`
}

type RedisConfig struct {
	Host     []string `json:"host" yaml:"host"`
	Port     int      `json:"port" yaml:"port"`
	Password string   `json:"password" yaml:"password"`
}

type KafkaWriter struct {
	Host     []string `json:"host" yaml:"host"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	Topic    []string `json:"topic" yaml:"topic"`
	GroupID  string   `json:"group_id" yaml:"group_id"`
}

type KafkaReader struct {
	Host     []string `json:"host" yaml:"host"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	Topic    []string `json:"topic" yaml:"topic"`
	GroupID  string   `json:"group_id" yaml:"group_id"`
}

type Server struct {
	Port int `json:"port" yaml:"port"`
}

type Config struct {
	Server      Server         `json:"server" yaml:"server"`
	Database    DatabaseConfig `json:"database" yaml:"database"`
	Redis       RedisConfig    `json:"redis" yaml:"redis"`
	KafkaWriter KafkaWriter    `json:"kafka_writer" yaml:"kafka_writer"`
	KafkaReader KafkaReader    `json:"kafka_reader" yaml:"kafka_reader"`
}

func Init() error {
	// 读取配置文件
	path := os.Getenv("CONFIG_PATH")
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			slog.Error("Failed to open config file", "error", err)
			return err
		}
		defer file.Close()

		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			if err := yaml.NewDecoder(file).Decode(Cfg); err != nil {
				return err
			}
		} else if strings.HasSuffix(path, ".json") {
			if err := json.NewDecoder(file).Decode(Cfg); err != nil {
				return err
			}
		} else {
			slog.Warn("Unsupported config file format, only .yaml, .yml and .json are supported")
		}
		slog.Info("Configuration loaded", "config", Cfg)
		return nil
	} else {
		slog.Warn("CONFIG_PATH environment variable is not set, using default configuration")
		return fmt.Errorf("CONFIG_PATH environment variable is not set")
	}
}
