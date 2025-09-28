package payment

type PaymentRequest struct {
	OrderID     uint64 `json:"order_id"`
	UserID      uint64 `json:"user_id"`
	Cost        int64  `json:"cost"`
	Description string `json:"description"`
	Attach      string `json:"attach"`
}
