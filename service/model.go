package service

type Order struct {
	OrderId       uint64  `json:"order_id" gorm:"primary_key;unique"`
	UserId        uint64  `json:"user_id" gorm:"index"`
	StreetAddress string  `json:"street_address"`
	City          string  `json:"city" gorm:"varchar(32)"`
	Phone         string  `json:"phone" gorm:"varchar(32)"`
	State         string  `json:"state" gorm:"varchar(32)"`
	ProductID     uint64  `json:"product_id"`
	Amount        uint64  `json:"amount"`
	Cost          float32 `json:"cost"`
	Status        int32   `json:"status" gorm:"type:smallint"`
}

type Orders struct {
	Items []Order `json:"items"`
}
