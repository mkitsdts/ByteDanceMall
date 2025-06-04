package config

// type RedisCluster struct {
// 	Host []string `json:"host"`
// 	Pass string	`json:"pass"`
// }

type RedisClient struct {
	Host []string `json:"host"`
}

type KafkaProducer struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

type KafkaReader struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
	GroupID string   `json:"group_id"`
}

type Configs struct {
	RedisConfig RedisClient `json:"redis_config"`
	// RedisConfig RedisCluster `json:"redis_config"`
	Producer KafkaProducer `json:"kafka_producer"`
	Reader   KafkaReader   `json:"kafka_reader"`
}
