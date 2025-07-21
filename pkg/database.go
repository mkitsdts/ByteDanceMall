package pkg

import (
	"bytedancemall/auth/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

func NewDatabase(cfg *config.Configs) (*Database, error) {
	masterDsn := cfg.MySQL.Master.Host + ":" + cfg.MySQL.Master.Port + "/" + cfg.MySQL.Master.Database +
		"?charset=utf8mb4&parseTime=True&loc=Local"
	masterDB, err := gorm.Open(mysql.Open(masterDsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	slaves := make([]*gorm.DB, len(cfg.MySQL.Slaves))
	for i, slave := range cfg.MySQL.Slaves {
		slaveDsn := slave.Host + ":" + slave.Port + "/" + slave.Database +
			"?charset=utf8mb4&parseTime=True&loc=Local"
		slaveDB, err := gorm.Open(mysql.Open(slaveDsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		slaves[i] = slaveDB
	}
	return &Database{
		Master: masterDB,
		Slaves: slaves,
	}, nil
}
