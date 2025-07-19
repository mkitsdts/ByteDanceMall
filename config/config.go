package config

import (
	"encoding/json"
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

type RedisConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RedisConfigs struct {
	Configs  []RedisConfig `json:"configs"`
	Password string        `json:"password"`
}

type AuthServerConfig struct {
	Address         string `json:"address"`
	DefaultUsername string `json:"default_username"`
}

type Config struct {
	MysqlConfig      MysqlConfigs     `json:"sql_config"`
	RedisConfig      RedisConfigs     `json:"redis_config"`
	AuthServerConfig AuthServerConfig `json:"auth_server_config"`
}

func InitConfigs() (*Config, error) {
	// 从配置文件中读取配置
	file, err := os.Open("configs.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	configs := &Config{}
	err = decoder.Decode(configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}
