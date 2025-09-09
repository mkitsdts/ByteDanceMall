package config

import (
	"os"

	"go.yaml.in/yaml/v2"
)

type RedisConfig struct {
	Addr     []string
	Password string
	DB       int
}

type Config struct {
	Redis RedisConfig
}

var Conf = &Config{
	Redis: RedisConfig{
		Addr:     []string{"localhost:6379"},
		Password: "",
		DB:       0,
	},
}

func Init() {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	if err = yaml.Unmarshal(file, Conf); err != nil {
		panic(err)
	}

}
