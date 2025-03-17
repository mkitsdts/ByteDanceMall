package model

type MysqlConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
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
	Configs []RedisConfig `json:"configs"`
	Password string `json:"password"`
}

type Configs struct {
	MysqlConfig MysqlConfigs `json:"sql_config"`
	RedisConfig RedisConfigs `json:"redis_config"`
}

type CartItem struct {
    UserId    uint32 `json:"id" gorm:"primaryKey;column:user_id"`
    ProductID uint32 `json:"product_id" gorm:"primaryKey"`
    Quantity  int32  `json:"quantity"`
}