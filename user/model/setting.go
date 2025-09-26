package model

type UserSettings struct {
	Id     uint64 `json:"id" gorm:"primary_key"`
	UserId uint64 `json:"user_id" gorm:"index"`
}
