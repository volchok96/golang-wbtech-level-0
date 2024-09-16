package models

type Order struct {
	ID                int
	OrderUid          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Locale            string   `json:"loacale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerID        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	Shardkey          string   `json:"shardkey"`
	SmID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Items  `json:"items"`
}