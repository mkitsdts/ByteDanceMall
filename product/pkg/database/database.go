package database

import (
	"bytedancemall/product/config"
	"bytedancemall/product/model"
	"fmt"
	"log/slog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

var db *Database

func Get() *gorm.DB {
	return db.Master
}

func Init() error {
	masterDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GetConfig().Mysql.User,
		config.GetConfig().Mysql.Password,
		config.GetConfig().Mysql.Master,
		config.GetConfig().Mysql.Port,
		config.GetConfig().Mysql.Database,
	)

	masterDB, err := gorm.Open(mysql.Open(masterDsn), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to master database", "error", err)
		return err
	}

	masterDB.AutoMigrate(&model.Product{})
	slog.Info("connected to master database successfully")

	slaves := make([]*gorm.DB, 0, len(config.GetConfig().Mysql.Slaves))
	for i := 0; i < len(config.GetConfig().Mysql.Slaves); i++ {
		slaveDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetConfig().Mysql.User,
			config.GetConfig().Mysql.Password,
			config.GetConfig().Mysql.Slaves[i],
			config.GetConfig().Mysql.Port,
			config.GetConfig().Mysql.Database,
		)
		slaveDB, err := gorm.Open(mysql.Open(slaveDsn), &gorm.Config{})
		if err != nil {
			slog.Error("failed to connect to slave database", "error", err)
			return err
		}
		slaves = append(slaves, slaveDB)
	}

	db = &Database{
		Master: masterDB,
		Slaves: slaves,
	}
	return nil
}
