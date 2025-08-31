package model

type DeductMessage struct {
	ProductId uint64 `json:"product_id"`
	Amount    uint64 `json:"amount"`
}
