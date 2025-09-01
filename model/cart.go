package model

// 购物车数据库表结构
type CartItem struct {
	UserId    uint64 `json:"id" gorm:"primaryKey;unique;column:user_id"`
	ProductID uint64 `json:"product_id" gorm:"index"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
	UpdatedAt string `json:"updated_at" gorm:"column:updated_at"`
}
