package model

// redis 存储结构

type Item struct {
	ProductID   uint64  `json:"product_id"`
	OriginPrice float64 `json:"origin_price"`
	Quantity    uint64  `json:"quantity"`
	Version     uint32  `json:"version"`
}

type Cart struct {
	Items []Item `json:"items"`
}
