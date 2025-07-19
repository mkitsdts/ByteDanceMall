package model

import "time"

type User struct {
	Id         uint64    `json:"id" gorm:"primary_key"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"password" gorm:"varchar(255),not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Status     int8      `json:"status" gorm:"default:0"`
	SettingsId uint64    `json:"settings_id" gorm:"default:0"`
}
