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

type Order struct {
	OrderId string `json:"order_id"`
	UserId uint32 `json:"user_id"`
	UserEmail string `json:"user_email"`
	UserCurrency string `json:"user_currency"`
	StreetAddress string `json:"street_address"`;
	City string `json:"city"`;
	ZipCode int32 `json:"zipcode"`;
	Phone string `json:"phone"`;
	ItemId uint32 `json:"item_id"`;
	Quantity int32 `json:"quantity"`;
	Cost float32 `json:"cost"`;
	Paid bool `json:"paid"`;
	State string `json:"state"`;
}