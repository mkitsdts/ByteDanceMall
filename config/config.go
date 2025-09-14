package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var Conf *Config = &Config{}

type DatabaseConfig struct {
	Master       string   `json:"master" yaml:"master"`
	Slaves       []string `json:"slaves" yaml:"slaves"`
	Name         string   `json:"name" yaml:"name"`
	Host         string   `json:"host" yaml:"host"`
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

type Server struct {
	Port int `json:"port" yaml:"port"`
}

type Config struct {
	Server   Server         `json:"server" yaml:"server"`
	Database DatabaseConfig `json:"database" yaml:"database"`
	Redis    RedisConfig    `json:"redis" yaml:"redis"`
}

func Init() error {
	file, err := os.Open("configs.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(Conf); err != nil {
		return err
	}

	return nil
}
