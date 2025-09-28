package config

var Cfg *Config = &Config{}

type DatabaseConfig struct {
	Master       string   `json:"master" yaml:"master"`
	Slaves       []string `json:"slaves" yaml:"slaves"`
	Name         string   `json:"name" yaml:"name"`
	Host         string   `json:"host" yaml:"host"`
	Port         int      `json:"port" yaml:"port"`
	Username     string   `json:"username" yaml:"username"`
	Password     string   `json:"password" yaml:"password"`
	MaxIdleConns int      `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int      `json:"max_open_conns" yaml:"max_open_conns"`
}

type RedisConfig struct {
	Host     []string `json:"host" yaml:"host"`
	Port     int      `json:"port" yaml:"port"`
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
	Host         string `json:"host" yaml:"host"`
	Port         int    `json:"port" yaml:"port"`
	MaxHTTPConns int    `json:"max_http_conns" yaml:"max_http_conns"`
}

type Etcd struct {
	Host      []string `json:"host" yaml:"host"`
	EnableTLS bool     `json:"tls" yaml:"tls"`
}

type WeChatPayment struct {
	AppID                      string `yaml:"app_id"`
	MchID                      string `yaml:"mch_id"`
	MchCertificateSerialNumber string `yaml:"mch_certificate_serial_number"`
	MchAPIv3Key                string `yaml:"mch_api_v3_key"`
	PrimaryKeyPath             string `yaml:"primary_key_path"`
	NotifyURL                  string `yaml:"notify_url"`
}

type Payment struct {
	WeChat WeChatPayment `yaml:"wechat"`
}

type Config struct {
	Etcd        Etcd           `json:"etcd" yaml:"etcd"`
	Server      Server         `json:"server" yaml:"server"`
	Database    DatabaseConfig `json:"database" yaml:"database"`
	Redis       RedisConfig    `json:"redis" yaml:"redis"`
	Payment     Payment        `json:"payment" yaml:"payment"`
	KafkaWriter KafkaWriter    `json:"kafka_writer" yaml:"kafka_writer"`
	KafkaReader KafkaReader    `json:"kafka_reader" yaml:"kafka_reader"`
}
