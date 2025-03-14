package model

type RedisConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RedisConfigs struct {
	Configs []RedisConfig `json:"configs"`
	Password string `json:"password"`
}

type Configs struct {
	RedisConfig RedisConfigs `json:"redis_config"`
}