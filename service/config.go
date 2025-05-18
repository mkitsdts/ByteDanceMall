package service

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type MysqlConfigs struct {
	Configs []MysqlConfig `json:"configs"`
}

type RedisConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RedisConfigs struct {
	Configs  []RedisConfig `json:"configs"`
	Password string        `json:"password"`
}

type AuthServerConfig struct {
	Address         string `json:"address"`
	DefaultUsername string `json:"default_username"`
}

type Configs struct {
	MysqlConfig      MysqlConfigs     `json:"sql_config"`
	RedisConfig      RedisConfigs     `json:"redis_config"`
	AuthServerConfig AuthServerConfig `json:"auth_server_config"`
}
