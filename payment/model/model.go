package model

import (
	"time"
)

type PaymentRecord struct {
	PaymentID uint64
}

type PaymentProcess struct {
	PaymentID    uint64
	CreatedAt    *time.Time
	PaySuccessAt *time.Time
	CanceledAt   *time.Time
	UpdatedAt    *time.Time
	CompletedAt  *time.Time
	Status       int32
}
