package config

// type RedisCluster struct {
// 	Host []string `json:"host"`
// 	Pass string	`json:"pass"`
// }

// type Configs struct {
// 	RedisConfig RedisCluster `json:"redis_config"`
// }

type RedisClient struct {
	Host []string `json:"host"`
	Pass string   `json:"pass"`
}

type Configs struct {
	RedisConfig RedisClient `json:"redis_config"`
}