package service

import (
	pb "bytedancemall/inventory/proto"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}
type InventoryService struct {
	DB     *Database
	Writer *kafka.Writer
	Reader map[string]*kafka.Reader
	Redis  *redis.ClusterClient
	pb.UnimplementedInventoryServiceServer
}

func NewInventoryService(db *Database, redis *redis.ClusterClient, writer *kafka.Writer, readers map[string]*kafka.Reader) *InventoryService {
	var service InventoryService
	service.DB = db
	service.Writer = writer
	service.Reader = readers
	service.Redis = redis
	return &service
}
