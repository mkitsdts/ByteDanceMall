package database

import (
	"bytedancemall/user/config"
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
	db := &Database{}

	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Master,
		config.Conf.Database.Port,
		config.Conf.Database.Name,
	)

	master, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	db.Master = master

	maxRetries := 5
	for _, m := range models {
		for i := range maxRetries {
			if err := migrate(master, m); err == nil {
				break
			}
			time.Sleep(10 << i * time.Millisecond)
			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to auto migrate model %v: %w", m, err)
			}
		}
	}

	if len(config.Conf.Database.Slaves) == 0 {
		slog.Info("database connected without replicas")
		return db, nil
	}

	var slaveDSNs []gorm.Dialector
	for _, slave := range config.Conf.Database.Slaves {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Conf.Database.Username,
			config.Conf.Database.Password,
			slave,
			config.Conf.Database.Port,
			config.Conf.Database.Name,
		)
		slaveDSNs = append(slaveDSNs, mysql.Open(slaveDSN))
	}

	if err := master.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterDSN)},
		Replicas: slaveDSNs,
		Policy:   dbresolver.RandomPolicy{},
	})); err != nil {
		return nil, err
	}

	if sqlDB, err := master.DB(); err == nil {
		sqlDB.SetMaxIdleConns(config.Conf.Database.MaxIdleConns)
		sqlDB.SetMaxOpenConns(config.Conf.Database.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	slog.Info("database connected successfully")
	return db, nil
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
