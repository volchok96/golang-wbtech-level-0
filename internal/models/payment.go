package models

type Payment struct {
	Transaction  string `json:"transaction"`
	ReqestID     string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTatal   int    `json:"goods_total"`
	Custom_fee   int    `json:"custom_fee"`
}