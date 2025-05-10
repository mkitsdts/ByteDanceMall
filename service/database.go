package service

import (
	"sync"

	"gorm.io/gorm"
)

type Database struct {
	Cluster   []*gorm.DB
	mu        sync.Mutex
	masterIdx int
	healthy   []bool
	Configs   []MysqlConfig
	ch        chan bool
}

func (db *Database) GetReader() *gorm.DB {
	db.mu.Lock()
	defer db.mu.Unlock()
	// 轮询获取健康的数据库连接
	for i := 0; i < len(db.Cluster); i++ {
		if db.healthy[db.masterIdx] {
			return db.Cluster[db.masterIdx]
		}
		db.masterIdx = (db.masterIdx + 1) % len(db.Cluster)
	}
	return nil
}
