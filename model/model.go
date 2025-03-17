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

type Product struct {
	ID          uint32    `json:"id" gorm:"primaryKey;autoIncrement"`
    Name        string    `json:"name" gorm:"type:varchar(255);index:idx_product_name;not null"`
    Description string    `json:"description" gorm:"type:text"`
    Price       float32   `json:"price" gorm:"type:decimal(10,2);not null;default:0.00"`
    ImageURL    string    `json:"image_url" gorm:"type:varchar(512)"`
    CategoryID  uint32    `json:"category_id" gorm:"index:idx_product_category"`
    Stock       int32     `json:"stock" gorm:"default:0;not null"`
    Status      uint8     `json:"status" gorm:"default:1;comment:'1:在售 2:下架 3:删除'"`
    Labels      string    `json:"labels" gorm:"type:varchar(255);comment:'标签，以逗号分隔'"`
}