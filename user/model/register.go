package model

import "time"

type Register struct {
	Id        uint64    `json:"id" gorm:"primary_key"`
	Username  string    `json:"username" gorm:"varchar(255)"`
	Email     string    `json:"email" gorm:"varchar(255)"`
	Password  string    `json:"password" gorm:"varchar(255),not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Status    int8      `json:"status" gorm:"default:0"`
}
