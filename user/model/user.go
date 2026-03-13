package model

import "time"

type User struct {
	Id           uint64    `json:"id" gorm:"primary_key"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Password     string    `json:"password" gorm:"varchar(255),not null"`
	PasswordSalt string    `json:"password_salt" gorm:"varchar(64),not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Status       int8      `json:"status" gorm:"default:0"`
	SettingsId   uint64    `json:"settings_id" gorm:"default:0"`
}

type LoginRecord struct {
	Id           uint64    `json:"id" gorm:"primary_key;autoIncrement"`
	UserId       uint64    `json:"user_id" gorm:"index;not null"`
	LoginAddress string    `json:"login_address" gorm:"varchar(255)"`
	LoginDevice  string    `json:"login_device" gorm:"varchar(255)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}
