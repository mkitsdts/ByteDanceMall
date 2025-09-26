package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MysqlConfig struct {
	Master   string   `json:"master" yaml:"master"`
	Slaves   []string `json:"slaves" yaml:"slaves"`
	Port     string   `json:"port" yaml:"port"`
	User     string   `json:"user" yaml:"user"`
	Password string   `json:"password" yaml:"password"`
	Database string   `json:"database" yaml:"database"`
}

type RedisConfig struct {
	Host     []string `json:"host" yaml:"host"`
	Port     string   `json:"port" yaml:"port"`
	Password string   `json:"password" yaml:"password"`
	DB       int      `json:"db" yaml:"db"`
}

type EsConfig struct {
	CloudID string `json:"cloud_id" yaml:"cloud_id"`
	APIKey  string `json:"api_key" yaml:"api_key"`
}

type Config struct {
	Mysql MysqlConfig `json:"mysql" yaml:"mysql"`
	Redis RedisConfig `json:"redis" yaml:"redis"`
	Es    EsConfig    `json:"es" yaml:"es"`
}

var conf *Config

func GetConfig() *Config {
	return conf
}

func Init() error {
	path := os.Getenv("CONFIG_PATH")
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	conf = &Config{}
	if err := decoder.Decode(conf); err != nil {
		return err
	}

	return nil
}
