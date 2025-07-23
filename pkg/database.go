package pkg

import (
	"bytedancemall/order/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	masterdsn := cfg.MysqlConfig.User + ":" + cfg.MysqlConfig.Password + "@tcp(" + cfg.MysqlConfig.Host[0] + ":" + cfg.MysqlConfig.Port + ")/" + cfg.MysqlConfig.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	masterdb, err := gorm.Open(mysql.Open(masterdsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	var slaves []*gorm.DB
	for _, host := range cfg.MysqlConfig.Host[1:] {
		slavedsn := cfg.MysqlConfig.User + ":" + cfg.MysqlConfig.Password + "@tcp(" + host + ":" + cfg.MysqlConfig.Port + ")/" + cfg.MysqlConfig.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
		slavedb, err := gorm.Open(mysql.Open(slavedsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		slaves = append(slaves, slavedb)
	}
	return &Database{
		Master: masterdb,
		Slaves: slaves,
	}, nil
}
