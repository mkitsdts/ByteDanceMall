package service

type Order struct {
	OrderId       string  `json:"order_id"`
	UserId        uint64  `json:"user_id"`
	StreetAddress string  `json:"street_address"`
	City          string  `json:"city" gorm:"varchar(32)"`
	Phone         string  `json:"phone" gorm:"varchar(32)"`
	ItemId        uint64  `json:"item_id"`
	Quantity      uint32  `json:"quantity"`
	Cost          float32 `json:"cost"`
	Paid          bool    `json:"paid"`
	State         string  `json:"state" gorm:"varchar(32)"`
}
