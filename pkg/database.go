package pkg

import (
	"bytedancemall/inventory/config"
	"bytedancemall/inventory/model"
	"fmt"
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

func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	var db Database
	// 连接主库
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Master,
		cfg.Port,
		cfg.Name,
	)

	master, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	// 准备从库DSN
	var slaveDSNs []gorm.Dialector
	for _, slave := range cfg.Slaves {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			slave,
			cfg.Port,
			cfg.Name,
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
		return nil, fmt.Errorf("failed to register dbresolver: %w", err)
	}

	// 配置连接池
	if sqlDB, err := master.DB(); err == nil {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// 自动迁移模型
	if err := master.AutoMigrate(&model.Inventory{}, &model.DevoteStock{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate models: %w", err)
	}

	return &db, nil
}
