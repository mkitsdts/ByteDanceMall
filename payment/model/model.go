package model

import (
	"time"
)

type PaymentOrder struct {
	OrderID   uint64 `gorm:"primaryKey"`
	UserID    uint64
	Method    string
	Status    int32
	Cost      int64
	OrderStr  *string    `gorm:"default:null"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
}

type PaymentRecord struct {
	ID        string     `gorm:"primaryKey;autoIncrement:true"`
	OrderID   uint64     `gorm:"index"`
	OrderStr  *string    `gorm:"default:null"`
	Status    int32      `gorm:"default:0"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
}
