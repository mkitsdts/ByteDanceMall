package pkg

import (
	"bytedancemall/inventory/config"
	"bytedancemall/inventory/model"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

func NewDatabase(models ...any) (*Database, error) {
	var db Database
	// 连接主库
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Cfg.Database.Username,
		config.Cfg.Database.Password,
		config.Cfg.Database.Host,
		config.Cfg.Database.Port,
		config.Cfg.Database.Name,
	)

	master, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return NewDatabase()
	}

	db.Master = master

	maxRetries := 5

	if len(config.Cfg.Database.Slaves) == 0 {

		for range maxRetries {
			if err := master.AutoMigrate(&model.Inventory{}); err == nil {
				return &db, nil
			}
			time.Sleep(time.Second * 2)
		}
	}

	// 准备从库DSN
	var slaveDSNs []gorm.Dialector
	for _, slave := range config.Cfg.Database.Slaves {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Cfg.Database.Username,
			config.Cfg.Database.Password,
			slave,
			config.Cfg.Database.Port,
			config.Cfg.Database.Name,
		)
		slaveDSNs = append(slaveDSNs, mysql.Open(slaveDSN))
	}

	// 使用DBResolver插件配置读写分离
	err = master.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterDSN)}, // 主库（写）
		Replicas: slaveDSNs,                               // 从库（读）
		Policy:   dbresolver.RandomPolicy{},               // 随机选择从库
	}))
	if err != nil {
		return NewDatabase()
	}

	// 配置连接池
	if sqlDB, err := master.DB(); err == nil {
		sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
		sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// 自动迁移模型
	for _, model := range models {
		for i := range maxRetries {
			if err := migrate(master, model); err == nil {
				break
			}
			time.Sleep(10 << i * time.Millisecond)
			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to auto migrate model %v: %w", model, err)
			}
		}
	}
	slog.Info("Database connected successfully")

	return &db, nil
}

func migrate(db *gorm.DB, model any) error {
	var err error
	for range 5 {
		if err = db.AutoMigrate(model); err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to auto migrate models: %w", err)
}
