package service

import (
	"bytedancemall/inventory/model"
	"bytedancemall/inventory/pkg"
	pb "bytedancemall/inventory/proto"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}
type InventoryService struct {
	deductScript *redis.Script
	DB           *Database
	Writer       map[string]*kafka.Writer
	Reader       map[string]*kafka.Reader
	// Redis  *redis.ClusterClient
	Redis *redis.Client
	pb.UnimplementedInventoryServiceServer
}

// func NewInventoryService(db *pkg.Database, redis *redis.ClusterClient, writer map[string]*kafka.Writer, readers map[string]*kafka.Reader) *InventoryService {
// 	var service InventoryService
// 	service.DB = &Database{
// 		Master: db.Master,
// 		Slaves: db.Slaves,
// 	}
// 	service.Writer = writer
// 	service.Reader = readers
// 	service.Redis = redis
// 	return &service
// }

func (s *InventoryService) LoopDeduct() {
	fmt.Println("Starting LoopDeduct...")
	for {
		msg, err := s.Reader["gomall-inventory-deduct"].FetchMessage(context.Background())
		if err != nil {
			slog.Info("No message", "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var dmsg model.DeductMessage
		if err := json.Unmarshal(msg.Value, &dmsg); err != nil {
			slog.Error("Failed to unmarshal message", "error", err)
			time.Sleep(100 * time.Millisecond)
			s.Reader["gomall-inventory-deduct"].CommitMessages(context.Background(), msg)
			continue
		}
		err = s.deduct(dmsg.ProductId, dmsg.Amount)
		if err != nil {
			slog.Error("Failed to deduct inventory", "error", err)
			continue
		}
		if err := s.Reader["gomall-inventory-deduct"].CommitMessages(context.Background(), msg); err != nil {
			slog.Error("Failed to commit message", "error", err)
		}
	}
}

func NewInventoryService(db *pkg.Database, rds *redis.Client, writer map[string]*kafka.Writer, readers map[string]*kafka.Reader) *InventoryService {
	var service InventoryService
	service.DB = &Database{
		Master: db.Master,
		Slaves: db.Slaves,
	}

	const deductLua = `
local product_key = KEYS[1]
local amount = tonumber(ARGV[1])
local current_stock = tonumber(redis.call("GET", product_key) or "-1")
if current_stock == -1 then
    return -1
end
if current_stock < amount then
    return -2
end
redis.call("DECRBY", product_key, amount)
return current_stock - amount
`

	service.Writer = writer
	service.Reader = readers
	service.Redis = rds
	service.deductScript = redis.NewScript(deductLua)

	go service.LoopDeduct()
	go service.LoopReduce()
	return &service
}
