package service

import (
	"bytedancemall/order/pkg"
	pb "bytedancemall/order/proto"

	"github.com/go-redsync/redsync/v4"                  // 引入 redsync 库，用于实现基于 Redis 的分布式锁
	"github.com/go-redsync/redsync/v4/redis/goredis/v9" // 引入 redsync 的 goredis 连接池
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

type OrderService struct {
	Db      *Database
	Redis   *redis.ClusterClient
	Writer  *kafka.Writer
	Reader  *kafka.Reader
	Redsync *redsync.Redsync
	pb.UnimplementedOrderServiceServer
}

// 创建一个新的订单服务实例
func NewOrderService(Db *pkg.Database, Redis *redis.ClusterClient, Writer *kafka.Writer, Reader *kafka.Reader) *OrderService {
	var s OrderService
	s.Db = &Database{
		Master: Db.Master,
		Slaves: Db.Slaves,
	}
	s.Redis = Redis
	s.Writer = Writer
	s.Reader = Reader
	pool := goredis.NewPool(Redis)
	s.Redsync = redsync.New(pool)
	return &s
}
