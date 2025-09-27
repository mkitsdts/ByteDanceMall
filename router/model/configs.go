package model

type MysqlConfig struct {
	User string `json:"user"`
	Password string `json:"password"`
	Host string `json:"host"`
	Port string `json:"port"`
	Database string `json:"database"`
}

type MysqlConfigs struct {
	Configs []MysqlConfig `json:"configs"`
}

type RedisConfig struct {
	Host string `json:"host"`
}

type RedisConfigs struct {
	Configs []RedisConfig `json:"configs"`
	Password string `json:"password"`
}

type AuthService struct {
	Host string `json:"host"`
}

type CartService struct {
	Host string `json:"host"`
}

type ProductService struct {
	Host string `json:"host"`
}

type OrderService struct {
	Host string `json:"host"`
}

type PaymentService struct {
	Host string `json:"host"`
}

type UserService struct {
	Host string `json:"host"`
}

type Configs struct {
	MysqlConfigs MysqlConfigs `json:"mysql"`
	RedisConfigs RedisConfigs `json:"redis"`
	AuthService AuthService `json:"auth_service"`
	CartService CartService `json:"cart_service"`
	ProductService ProductService `json:"product_service"`
	OrderService OrderService `json:"order_service"`
	PaymentService PaymentService `json:"payment_service"`
	UserService UserService `json:"user_service"`
}