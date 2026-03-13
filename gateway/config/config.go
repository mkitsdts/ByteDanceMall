package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type AuthService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type CartService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type InventoryService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type OrderService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type PaymentService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type ProductService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type UserService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LLMService struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Server struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	LimiterRate   int    `yaml:"limiter_rate"`
	LimiterWindow int    `yaml:"limiter_window"`
}

type Config struct {
	Auth      AuthService      `yaml:"auth"`
	Cart      CartService      `yaml:"cart"`
	Inventory InventoryService `yaml:"inventory"`
	Order     OrderService     `yaml:"order"`
	Payment   PaymentService   `yaml:"payment"`
	Product   ProductService   `yaml:"product"`
	User      UserService      `yaml:"user"`
	LLM       LLMService       `yaml:"llm"`
	Server    Server           `yaml:"server"`
}

var conf *Config

func Get() *Config {
	return conf
}

func init() {
	// Initialize the config
	file, err := os.Open("config.yaml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&conf); err != nil {
		panic(err)
	}
	slog.Info("config loaded", slog.Any("config", conf))
}
