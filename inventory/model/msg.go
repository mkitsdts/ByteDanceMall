package model

type DeductMessage struct {
	OrderId   uint64 `json:"order_id"`
	ProductId uint64 `json:"product_id"`
	Amount    uint64 `json:"amount"`
}
