package model

// 商品数据库表结构
type ProductItem struct {
	ProductID uint64  `json:"product_id" gorm:"primaryKey;index"`
	Price     float64 `json:"price" gorm:"column:price"`
	CreatedAt uint64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt uint64  `json:"updated_at" gorm:"column:updated_at"`
}

// 购物车数据库表结构
type CartItem struct {
	ID          uint64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint64  `json:"user_id" gorm:"not null;uniqueIndex:uniq_user_product,priority:1;index"`
	ProductID   uint64  `json:"product_id" gorm:"not null;uniqueIndex:uniq_user_product,priority:2"`
	OriginPrice float64 `json:"origin_price" gorm:"column:origin_price"`
	Quantity    uint32  `json:"quantity" gorm:"column:quantity"`
	CreatedAt   uint64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   uint64  `json:"updated_at" gorm:"column:updated_at"`
}
