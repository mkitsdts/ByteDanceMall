package database

import (
	"bytedancemall/order/config"
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

var db *Database = &Database{}

func DB() *gorm.DB {
	return db.Master
}

func NewDatabase(models ...any) error {
	// 连接主库
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Cfg.Database.Username,
		config.Cfg.Database.Password,
		config.Cfg.Database.Master,
		config.Cfg.Database.Port,
		config.Cfg.Database.Name,
	)
	maxRetries := 5
	master, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		slog.Error("Failed to connect to master database", " dsn", masterDSN)
		return NewDatabase()
	}

	for _, model := range models {
		for i := range maxRetries {
			if err := migrate(master, model); err == nil {
				break
			}
			time.Sleep(10 << i * time.Millisecond)
			if i == maxRetries-1 {
				return fmt.Errorf("failed to auto migrate model %v: %w", model, err)
			}
		}
	}
	slog.Info("Database connected successfully")
	db.Master = master

	if len(config.Cfg.Database.Slaves) == 0 {
		slog.Info("No slave databases configured, skipping read-write splitting")
		return nil
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
		time.Sleep(100 * time.Millisecond)
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
				return fmt.Errorf("failed to auto migrate model %v: %w", model, err)
			}
		}
	}
	slog.Info("Database connected successfully")

	return nil
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
