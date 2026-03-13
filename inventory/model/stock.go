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
	RecordID    string     `json:"record_id" gorm:"primaryKey;size:64"`
	ProductID   uint64     `json:"product_id" gorm:"index:idx_out_inventory_product_state_bucket,priority:1;index"`
	OrderId     *uint64    `json:"order_id" gorm:"index"`
	Amount      uint64     `json:"amount"`
	Bucket      int32      `json:"bucket" gorm:"index:idx_out_inventory_product_state_bucket,priority:3"`
	State       int8       `json:"state" gorm:"index:idx_out_inventory_product_state_bucket,priority:2"`
	CommittedAt *time.Time `json:"committed_at"`
	CreatedAt   *time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
