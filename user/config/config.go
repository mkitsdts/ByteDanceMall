package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type MysqlConfigs struct {
	Configs []MysqlConfig `json:"configs"`
}

type RedisConfigs struct {
	Host     []string `json:"host"`
	Password string   `json:"password"`
}

type Config struct {
	MysqlConfig MysqlConfigs `json:"database"`
	RedisConfig RedisConfigs `json:"redis"`
}

var Cfg *Config

func InitConfigs(path string) error {
	// 从配置文件中读取配置
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	Cfg = &Config{}
	err = decoder.Decode(Cfg)
	if err != nil {
		return err
	}

	return nil
}
