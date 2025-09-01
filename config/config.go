package config

import (
	"encoding/json"
	"os"
)

var Cfg *Config = &Config{}

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

type Server struct {
	Port int `json:"port"`
}

type Config struct {
	Server   Server         `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
}

func Init() error {
	file, err := os.Open("configs.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(Cfg); err != nil {
		return err
	}

	return nil
}
