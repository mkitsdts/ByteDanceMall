package model

import (
	"time"
)

type PaymentRecord struct {
	PaymentID string
	OrderID   uint64
	UserID    uint64
	Method    string
	Status    int32
	Cost      int64
	OrderStr  *string    `gorm:"default:null"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
}

type PaymentProcess struct {
	PaymentID    string
	CreatedAt    *time.Time `gorm:"autoCreateTime"`
	PaySuccessAt *time.Time `gorm:"default:null"`
	CanceledAt   *time.Time `gorm:"default:null"`
	UpdatedAt    *time.Time `gorm:"default:null"`
	CompletedAt  *time.Time `gorm:"default:null"`
	Description  string
}
