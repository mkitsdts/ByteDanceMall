package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RedisConfig struct {
	Host     []string `yaml:"host"`
	Port     int      `yaml:"port"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
}

type ModelConfig struct {
	Enable bool   `yaml:"enable"`
	Name   string `yaml:"name"`
	Host   string `yaml:"host"`
	Key    string `yaml:"key"`
}

type Config struct {
	Redis RedisConfig `yaml:"redis"`
	LLM   ModelConfig `yaml:"llm"`
	Embed ModelConfig `yaml:"embed"`
}

var Conf = &Config{}

func Init() {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(file, Conf)
	if err != nil {
		panic(err)
	}
}
