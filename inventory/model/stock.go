package model

import (
	"time"

	"gorm.io/gorm"
)

type Inventory struct {
	ProductID   uint64         `json:"product_id" gorm:"primaryKey;unique"`
	TotalStock  uint64         `json:"total_stock"`
	LockedStock uint64         `json:"locked_stock"`
	State       int8           `json:"state"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"` // Optional field for soft deletion
	Version     int64          `json:"version"`                 // Optimistic locking version
}

type OutInventory struct {
	ProductID uint64     `json:"product_id"`
	OrderId   uint64     `json:"order_id" gorm:"primaryKey;unique"`
	Amount    uint64     `json:"amount"`
	State     int8       `json:"state"`
	CreatedAt *time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
