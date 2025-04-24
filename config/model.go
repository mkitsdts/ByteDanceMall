package config

// type RedisCluster struct {
// 	Host []string `json:"host"`
// 	Pass string	`json:"pass"`
// }

type RedisClient struct {
	Host []string `json:"host"`
	Pass string   `json:"pass"`
}

type KafkaProducer struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

type Configs struct {
	RedisConfig RedisClient `json:"redis_config"`
	// RedisConfig RedisCluster `json:"redis_config"`
	KafkaConfig KafkaProducer `json:"kafka_config"`
}
