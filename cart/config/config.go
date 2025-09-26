package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

var Cfg *Config = &Config{}

type DatabaseConfig struct {
	Master       string   `json:"master" yaml:"master"`
	Slaves       []string `json:"slaves" yaml:"slaves"`
	Name         string   `json:"name" yaml:"name"`
	Host         string   `json:"host" yaml:"host"`
	Port         string   `json:"port" yaml:"port"`
	Username     string   `json:"username" yaml:"username"`
	Password     string   `json:"password" yaml:"password"`
	MaxIdleConns int      `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int      `json:"max_open_conns" yaml:"max_open_conns"`
}

type RedisConfig struct {
	Host     []string `json:"host" yaml:"host"`
	Port     string   `json:"port" yaml:"port"`
	Password string   `json:"password" yaml:"password"`
}

type KafkaWriter struct {
	Host     []string `json:"host" yaml:"host"`
	Port     string   `json:"port" yaml:"port"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	Topic    []string `json:"topic" yaml:"topic"`
	GroupID  string   `json:"group_id" yaml:"group_id"`
}

type KafkaReader struct {
	Host     []string `json:"host" yaml:"host"`
	Port     string   `json:"port" yaml:"port"`
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

	if os.Getenv("CONFIG_PATH") != "" {
		file, err := os.Open(os.Getenv("CONFIG_PATH"))
		if err == nil {
			decoder := yaml.NewDecoder(file)
			if err := decoder.Decode(&Cfg); err == nil {
				file.Close()
				return nil
			}
			file.Close()
			return err
		}
		file.Close()
	}
	if os.Getenv("MYSQL_DATABASE") != "" {
		Cfg.Database.Name = os.Getenv("MYSQL_DATABASE")
	} else {
		Cfg.Database.Name = "cart"
	}

	if os.Getenv("MYSQL_HOST") != "" {
		Cfg.Database.Host = os.Getenv("MYSQL_HOST")
	} else {
		Cfg.Database.Host = "localhost"
	}

	if os.Getenv("MYSQL_PORT") != "" {
		Cfg.Database.Port = os.Getenv("MYSQL_PORT")
	} else {
		Cfg.Database.Port = "3306"
	}

	if os.Getenv("MYSQL_USER") != "" {
		Cfg.Database.Username = os.Getenv("MYSQL_USER")
	} else {
		Cfg.Database.Username = "user"
	}

	if os.Getenv("MYSQL_PASSWORD") != "" {
		Cfg.Database.Password = os.Getenv("MYSQL_PASSWORD")
	} else {
		Cfg.Database.Password = "password"
	}

	if os.Getenv("REDIS_HOST") != "" {
		Cfg.Redis.Host = []string{os.Getenv("REDIS_HOST")}
	} else {
		Cfg.Redis.Host = []string{"localhost"}
	}

	if os.Getenv("REDIS_PORT") != "" {
		Cfg.Redis.Port = os.Getenv("REDIS_PORT")
	} else {
		Cfg.Redis.Port = "6379"
	}

	if os.Getenv("REDIS_PASSWORD") != "" {
		Cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	} else {
		Cfg.Redis.Password = "password"
	}

	if os.Getenv("KAFKA_HOST") != "" {
		Cfg.KafkaWriter.Host = []string{os.Getenv("KAFKA_HOST")}
		Cfg.KafkaReader.Host = []string{os.Getenv("KAFKA_HOST")}
	} else {
		Cfg.KafkaWriter.Host = []string{"localhost"}
		Cfg.KafkaReader.Host = []string{"localhost"}
	}

	if os.Getenv("KAFKA_PORT") != "" {
		Cfg.KafkaWriter.Port = os.Getenv("KAFKA_PORT")
		Cfg.KafkaReader.Port = os.Getenv("KAFKA_PORT")
	} else {
		Cfg.KafkaWriter.Port = "9092"
		Cfg.KafkaReader.Port = "9092"
	}

	if os.Getenv("KAFKA_USERNAME") != "" {
		Cfg.KafkaWriter.Username = os.Getenv("KAFKA_USERNAME")
		Cfg.KafkaReader.Username = os.Getenv("KAFKA_USERNAME")
	} else {
		Cfg.KafkaWriter.Username = "user"
		Cfg.KafkaReader.Username = "user"
	}

	if os.Getenv("KAFKA_PASSWORD") != "" {
		Cfg.KafkaWriter.Password = os.Getenv("KAFKA_PASSWORD")
		Cfg.KafkaReader.Password = os.Getenv("KAFKA_PASSWORD")
	} else {
		Cfg.KafkaWriter.Password = "password"
		Cfg.KafkaReader.Password = "password"
	}

	if os.Getenv("KAFKA_TOPIC") != "" {
		Cfg.KafkaWriter.Topic = []string{os.Getenv("KAFKA_TOPIC")}
		Cfg.KafkaReader.Topic = []string{os.Getenv("KAFKA_TOPIC")}
	} else {
		Cfg.KafkaWriter.Topic = []string{"cart_topic"}
		Cfg.KafkaReader.Topic = []string{"cart_topic"}
	}

	if os.Getenv("KAFKA_GROUP_ID") != "" {
		Cfg.KafkaWriter.GroupID = os.Getenv("KAFKA_GROUP_ID")
		Cfg.KafkaReader.GroupID = os.Getenv("KAFKA_GROUP_ID")
	} else {
		Cfg.KafkaWriter.GroupID = "cart_group"
		Cfg.KafkaReader.GroupID = "cart_group"
	}

	if os.Getenv("SERVER_PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
		if err == nil {
			Cfg.Server.Port = port
		}
	} else {
		Cfg.Server.Port = 14803
	}

	return nil
}
