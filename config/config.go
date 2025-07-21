package config

import (
	"encoding/json"
	"os"
)

type RedisConfig struct {
	Host     []string `json:"Host"`
	Password string   `json:"password"`
	Port     string   `json:"port"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type MySQLConfigs struct {
	Master MySQLConfig   `json:"master"`
	Slaves []MySQLConfig `json:"slaves"`
}

type Configs struct {
	Redis RedisConfig  `json:"redis_config"`
	MySQL MySQLConfigs `json:"mysql_config"`
}

func NewConfig() (*Configs, error) {
	file, err := os.Open("configs.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	configs := &Configs{}
	err = decoder.Decode(configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}
