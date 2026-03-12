package model

type UserBlacklist struct {
	UserID    uint64 `gorm:"column:user_id;primaryKey" json:"user_id"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}
