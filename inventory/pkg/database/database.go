package database

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

func New(models ...any) (*Database, error) {
	var db Database
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
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
		time.Sleep(time.Second * 5)
		return New(models...)
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

	var slaveDSNs []gorm.Dialector
	for _, slave := range config.Cfg.Database.Slaves {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Cfg.Database.Username,
			config.Cfg.Database.Password,
			slave,
			config.Cfg.Database.Port,
			config.Cfg.Database.Name,
		)
		slaveDSNs = append(slaveDSNs, mysql.Open(slaveDSN))
	}

	err = master.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterDSN)},
		Replicas: slaveDSNs,
		Policy:   dbresolver.RandomPolicy{},
	}))
	if err != nil {
		return New(models...)
	}

	if sqlDB, err := master.DB(); err == nil {
		sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
		sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

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
