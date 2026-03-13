package service

import (
	pb "bytedancemall/inventory/proto"
	"bytedancemall/inventory/usecase"

	"github.com/segmentio/kafka-go"
)

const (
	inventoryRecoveryTopic = "gomall-inventory-recovery"
	inventoryCommitTopic   = "gomall-inventory-reduce"
)

type InventoryService struct {
	usecase *usecase.InventoryUsecase
	reader  map[string]*kafka.Reader
	pb.UnimplementedInventoryServiceServer
}

func NewInventoryService(uc *usecase.InventoryUsecase, readers map[string]*kafka.Reader) *InventoryService {
	svc := &InventoryService{
		usecase: uc,
		reader:  readers,
	}
	svc.startConsumers()
	return svc
}

func availableStock(total, locked uint64) uint64 {
	if total < locked {
		return 0
	}
	return total - locked
}
