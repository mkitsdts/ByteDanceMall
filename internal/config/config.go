package config

import "github.com/spf13/viper"

type RedisCluster struct {
	Host []string `json:"host"`
}

type KafkaProducer struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

type GrpcConfig struct {
	Address string `json:"address"`
}

type JaegerConfig struct {
	Endpoint    string `json:"endpoint"`
	ServiceName string `json:"service_name"`
}

type Config struct {
	Redis  RedisCluster  `json:"redis"`
	Kafka  KafkaProducer `json:"kafka"`
	GRPC   GrpcConfig    `json:"grpc"`
	Jaeger JaegerConfig  `json:"jaeger"`
}

// NewConfig 创建配置
func NewConfig() (*Config, error) {
	viper.SetConfigName("configs")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("grpc.address", ":50051")
	viper.SetDefault("jaeger.service_name", "seckill-service")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
