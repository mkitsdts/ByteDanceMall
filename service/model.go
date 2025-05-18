package service

import "time"

type User struct {
	Id         int64     `json:"id" gorm:"primary_key"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"password" gorm:"varchar(255),not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Status     int8      `json:"status" gorm:"default:0"`
	SettingsId int64     `json:"settings_id" gorm:"default:0"`
}

type Register struct {
	Id        int64     `json:"id" gorm:"primary_key"`
	Username  string    `json:"username" gorm:"varchar(255)"`
	Email     string    `json:"email" gorm:"varchar(255)"`
	Password  string    `json:"password" gorm:"varchar(255),not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Status    int8      `json:"status" gorm:"default:0"`
}

type UserSettings struct {
	Id     int64 `json:"id" gorm:"primary_key"`
	UserId int64 `json:"user_id" gorm:"index"`
}
