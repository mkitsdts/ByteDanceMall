package model

type Value struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
}

type PaymentMsg struct {
	OrderID string `json:"key"`
	Value   Value  `json:"value"`
}
