package model

type RefreshToken struct {
	UserID    uint64 `gorm:"column:user_id;primaryKey" json:"user_id"`
	Token     string `gorm:"column:token;size:256;not null;uniqueIndex" json:"token"`
	Expiry    int64  `gorm:"column:expiry;not null" json:"expiry"`   // Unix 时间戳，秒级
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"` // 毫秒级
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"` // 毫秒级
}
