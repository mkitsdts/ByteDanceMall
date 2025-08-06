package model

import "time"

type Inventory struct {
	ProductID   uint64    `json:"product_id"`
	TotalStock  uint64    `json:"total_stock"`
	LockedStock uint64    `json:"locked_stock"`
	State       int8      `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"` // Optional field for soft deletion
	Version     int64     `json:"version"`    // Optimistic locking version
}

type DevoteStock struct {
	ProductID uint64    `json:"product_id"`
	Amount    uint64    `json:"amount"`
	State     int8      `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"` // Optional field for soft deletion
	Version   int64     `json:"version"`    // Optimistic locking version
}
