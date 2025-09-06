package models

import "time"

type Order struct {
	OrderUID          string    `json:"order_uid" gorm:"primaryKey;size:255;not null"`
	TrackNumber       string    `json:"track_number" gorm:"size:255;not null"`
	Entry             string    `json:"entry" gorm:"size:10"`
	Delivery          Delivery  `json:"delivery" gorm:"embedded;embeddedPrefix:delivery_"`
	Payment           Payment   `json:"payment" gorm:"embedded;embeddedPrefix:payment_"`
	Items             []Item    `json:"items" gorm:"foreignKey:OrderUID;references:OrderUID"`
	Locale            string    `json:"locale" gorm:"size:10"`
	InternalSignature string    `json:"internal_signature" gorm:"size:500"`
	CustomerID        string    `json:"customer_id" gorm:"size:255;not null"`
	DeliveryService   string    `json:"delivery_service" gorm:"size:255"`
	Shardkey          string    `json:"shardkey" gorm:"size:10"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created" gorm:"not null"`
	OofShard          string    `json:"oof_shard" gorm:"size:10"`
}

type Delivery struct {
	Name    string `json:"name" gorm:"size:255;not null"`
	Phone   string `json:"phone" gorm:"size:20;not null"`
	Zip     string `json:"zip" gorm:"size:20"`
	City    string `json:"city" gorm:"size:100;not null"`
	Address string `json:"address" gorm:"size:500;not null"`
	Region  string `json:"region" gorm:"size:100"`
	Email   string `json:"email" gorm:"size:255"`
}

type Payment struct {
	Transaction  string `json:"transaction" gorm:"size:255;not null"`
	RequestID    string `json:"request_id" gorm:"size:255"`
	Currency     string `json:"currency" gorm:"size:10;not null"`
	Provider     string `json:"provider" gorm:"size:100;not null"`
	Amount       int    `json:"amount" gorm:"not null"`
	PaymentDt    int    `json:"payment_dt" gorm:"not null"`
	Bank         string `json:"bank" gorm:"size:100"`
	DeliveryCost int    `json:"delivery_cost" gorm:"not null"`
	GoodsTotal   int    `json:"goods_total" gorm:"not null"`
	CustomFee    int    `json:"custom_fee" gorm:"default:0"`
}

type Item struct {
	ID          uint   `json:"-" gorm:"primaryKey;autoIncrement"`
	OrderUID    string `json:"-" gorm:"size:255;not null;index"`
	ChrtID      int    `json:"chrt_id" gorm:"not null"`
	TrackNumber string `json:"track_number" gorm:"size:255;not null"`
	Price       int    `json:"price" gorm:"not null"`
	Rid         string `json:"rid" gorm:"size:255;not null"`
	Name        string `json:"name" gorm:"size:500;not null"`
	Sale        int    `json:"sale"`
	Size        string `json:"size" gorm:"size:50"`
	TotalPrice  int    `json:"total_price" gorm:"not null"`
	NmID        int    `json:"nm_id" gorm:"not null"`
	Brand       string `json:"brand" gorm:"size:255"`
	Status      int    `json:"status" gorm:"not null"`
}
