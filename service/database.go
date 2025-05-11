package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Database struct {
	Client      *gorm.DB     // 使用单一的gorm.DB实例，插件会处理读写分离
	mu          sync.RWMutex // 使用读写锁提高性能
	ch          chan bool
	configs     []MysqlConfig      // 保存配置用于重连
	stopChecker context.CancelFunc // 停止健康检查
}

func (db *Database) InitDatabase(configs []MysqlConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("no database configs provided")
	}

	// 保存配置副本用于重连
	db.configs = make([]MysqlConfig, len(configs))
	copy(db.configs, configs)

	// 构建主库连接字符串
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		configs[0].User, configs[0].Password, configs[0].Host, configs[0].Port, configs[0].Database)

	// 打开到主库的连接
	var err error
	db.Client, err = gorm.Open(mysql.Open(masterDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to master database: %w", err)
	}

	// 如果只有一个数据库配置，不需要使用DBResolver
	if len(configs) == 1 {
		return nil
	}

	// 构建从库连接
	var replicas []gorm.Dialector
	for i := 1; i < len(configs); i++ {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			configs[i].User, configs[i].Password, configs[i].Host, configs[i].Port, configs[i].Database)
		replicas = append(replicas, mysql.Open(slaveDSN))
	}

	// 配置DBResolver插件
	err = db.Client.Use(dbresolver.Register(dbresolver.Config{
		Replicas:          replicas,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}).SetMaxIdleConns(10).
		SetMaxOpenConns(100).
		SetConnMaxLifetime(time.Hour))

	if err != nil {
		return fmt.Errorf("failed to configure DBResolver: %w", err)
	}

	// 启动健康检查和故障转移协程
	ctx, cancel := context.WithCancel(context.Background())
	db.stopChecker = cancel
	go db.healthCheck(ctx)

	return nil
}

// 健康检查和故障转移
func (db *Database) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			db.checkMaster()
			// 可以增加从库健康检查
		}
	}
}

// 主库健康检查
func (db *Database) checkMaster() {
	db.mu.RLock()
	masterDB := db.Client
	db.mu.RUnlock()

	sqlDB, err := masterDB.DB()
	if err != nil {
		log.Printf("Warning: Cannot get sql.DB from master: %v", err)
		db.triggerFailover()
		return
	}

	// 使用超时上下文进行健康检查
	checkCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = sqlDB.PingContext(checkCtx)
	if err != nil {
		log.Printf("Warning: Master database health check failed: %v", err)
		db.triggerFailover()
	}
}

// 触发故障转移
func (db *Database) triggerFailover() {
	db.mu.Lock()
	defer db.mu.Unlock()

	log.Println("Triggering database failover")

	// 首先尝试重连主库
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.configs[0].User, db.configs[0].Password,
		db.configs[0].Host, db.configs[0].Port, db.configs[0].Database)

	newMasterDB, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{})
	if err == nil {
		// 主库恢复连接
		sqlDB, err := newMasterDB.DB()
		if err == nil && sqlDB.Ping() == nil {
			log.Println("Master database reconnected successfully")
			db.Client = newMasterDB
			// 重新配置从库
			db.reconfigureReplicas()
			return
		}
	}

	log.Println("Master database reconnection failed, promoting replica")

	// 主库重连失败，需要提升从库为主库
	// 寻找健康的从库
	for i := 1; i < len(db.configs); i++ {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.configs[i].User, db.configs[i].Password,
			db.configs[i].Host, db.configs[i].Port, db.configs[i].Database)

		newMasterDB, err := gorm.Open(mysql.Open(slaveDSN), &gorm.Config{})
		if err != nil {
			continue
		}

		sqlDB, err := newMasterDB.DB()
		if err != nil || sqlDB.Ping() != nil {
			continue
		}

		// 找到健康从库，将其提升为主库
		log.Printf("Promoted replica %d to master", i)

		// 交换配置，使当前从库成为主库
		db.configs[0], db.configs[i] = db.configs[i], db.configs[0]

		// 设置新主库
		db.Client = newMasterDB

		// 重新配置其他从库
		db.reconfigureReplicas()
		return
	}

	log.Println("CRITICAL: All database nodes are down!")
}

// 重新配置从库
func (db *Database) reconfigureReplicas() {
	var replicas []gorm.Dialector
	for i := 1; i < len(db.configs); i++ {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.configs[i].User, db.configs[i].Password,
			db.configs[i].Host, db.configs[i].Port, db.configs[i].Database)
		replicas = append(replicas, mysql.Open(slaveDSN))
	}

	// 更新DBResolver配置
	err := db.Client.Use(dbresolver.Register(dbresolver.Config{
		Replicas:          replicas,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}).SetMaxIdleConns(10).
		SetMaxOpenConns(100).
		SetConnMaxLifetime(time.Hour))

	if err != nil {
		log.Printf("Error reconfiguring replicas: %v", err)
	}
}

// GetDB 返回数据库连接
func (db *Database) Get() *gorm.DB {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.Client
}

// GetReader 获取读库连接（自动处理负载均衡）
func (db *Database) GetReader() *gorm.DB {
	// DBResolver会自动选择从库执行查询
	return db.Get().Clauses(dbresolver.Read)
}

// GetWriter 获取写库连接
func (db *Database) GetWriter() *gorm.DB {
	// 确保写操作在主库执行
	return db.Get().Clauses(dbresolver.Write)
}

// Close 关闭数据库连接
func (db *Database) Close() {
	// 停止健康检查
	if db.stopChecker != nil {
		db.stopChecker()
	}

	db.mu.Lock()
	defer db.mu.Unlock()
	if db.Client != nil {
		sqlDB, err := db.Client.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// 关闭通知通道
	if db.ch != nil {
		close(db.ch)
	}
}
