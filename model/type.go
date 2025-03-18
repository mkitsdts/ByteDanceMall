package model

type CartItem struct {
	ProductId uint32 `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}

type Address struct{
  StreetAddress string `json:"street_address"`;
  City string `json:"city"`;
  State string `json:"state"`;
  Country string `json:"country"`;
  ZipCode int32 `json:"zip_code"`;
}

type OrderItem struct {
	ProductId uint32 `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Cost float32 `json:"cost"`;
}

type Order struct {
	Items []OrderItem `json:"items"`
	Address Address `json:"address"`;
	UserCurrency string `json:"user_currency"`;
	UserId uint32 `json:"user_id"`;
	Email string `json:"email"`;
}

type CreditCardInfo struct {
	CreditCardNumber string `json:"credit_card_number"`;
	CreditCardCvv int32 `json:"credit_card_cvv"`;
	CreditCardExpirationYear int32 `json:"credit_card_expiration_year"`;
	CreditCardExpirationMonth int32 `json:"credit_card_expiration_month"`;
}

type Payment struct {
	OrderId string `json:"order_id"`;
	Amount float32 `json:"amount"`;
	CreditCard CreditCardInfo `json:"credit_card"`;
	UserId uint32 `json:"user_id"`;
}

type Page struct {
	Page int32 `json:"page"`
	PageSize int64 `json:"page_size"`
	CategoryName string `json:"category_name"`
}

type Product struct {

}

type SearchRequest struct {
	Query string `json:"query"`;
}