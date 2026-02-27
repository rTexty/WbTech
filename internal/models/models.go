// Package models defines the data structures used in the application.
package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Order represents the main order structure.
type Order struct {
	OrderUID          string    `json:"order_uid" gorm:"primaryKey;size:255;not null" validate:"required"`
	TrackNumber       string    `json:"track_number" gorm:"size:255;not null" validate:"required"`
	Entry             string    `json:"entry" gorm:"size:10" validate:"required"`
	Delivery          Delivery  `json:"delivery" gorm:"embedded;embeddedPrefix:delivery_" validate:"required"`
	Payment           Payment   `json:"payment" gorm:"embedded;embeddedPrefix:payment_" validate:"required"`
	Items             []Item    `json:"items" gorm:"foreignKey:OrderUID;references:OrderUID" validate:"required,dive"`
	Locale            string    `json:"locale" gorm:"size:10" validate:"required"`
	InternalSignature string    `json:"internal_signature" gorm:"size:500"`
	CustomerID        string    `json:"customer_id" gorm:"size:255;not null" validate:"required"`
	DeliveryService   string    `json:"delivery_service" gorm:"size:255" validate:"required"`
	Shardkey          string    `json:"shardkey" gorm:"size:10" validate:"required"`
	SmID              int       `json:"sm_id" validate:"required"`
	DateCreated       time.Time `json:"date_created" gorm:"not null" validate:"required"`
	OofShard          string    `json:"oof_shard" gorm:"size:10" validate:"required"`
}

// Validate checks the structural integrity of the Order.
func (o *Order) Validate() error {
	return validate.Struct(o)
}

// Delivery contains delivery information.
type Delivery struct {
	Name    string `json:"name" gorm:"size:255;not null" validate:"required"`
	Phone   string `json:"phone" gorm:"size:20;not null" validate:"required"`
	Zip     string `json:"zip" gorm:"size:20" validate:"required"`
	City    string `json:"city" gorm:"size:100;not null" validate:"required"`
	Address string `json:"address" gorm:"size:500;not null" validate:"required"`
	Region  string `json:"region" gorm:"size:100" validate:"required"`
	Email   string `json:"email" gorm:"size:255" validate:"required,email"`
}

// Payment contains payment details.
type Payment struct {
	Transaction  string `json:"transaction" gorm:"size:255;not null" validate:"required"`
	RequestID    string `json:"request_id" gorm:"size:255"`
	Currency     string `json:"currency" gorm:"size:10;not null" validate:"required"`
	Provider     string `json:"provider" gorm:"size:100;not null" validate:"required"`
	Amount       int    `json:"amount" gorm:"not null" validate:"required,gte=0"`
	PaymentDt    int    `json:"payment_dt" gorm:"not null" validate:"required"`
	Bank         string `json:"bank" gorm:"size:100" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" gorm:"not null" validate:"required,gte=0"`
	GoodsTotal   int    `json:"goods_total" gorm:"not null" validate:"required,gte=0"`
	CustomFee    int    `json:"custom_fee" gorm:"default:0" validate:"gte=0"`
}

// Item represents an item in the order.
type Item struct {
	ID          uint   `json:"-" gorm:"primaryKey;autoIncrement"`
	OrderUID    string `json:"-" gorm:"size:255;not null;index"`
	ChrtID      int    `json:"chrt_id" gorm:"not null" validate:"required"`
	TrackNumber string `json:"track_number" gorm:"size:255;not null" validate:"required"`
	Price       int    `json:"price" gorm:"not null" validate:"required,gte=0"`
	Rid         string `json:"rid" gorm:"size:255;not null" validate:"required"`
	Name        string `json:"name" gorm:"size:500;not null" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0"`
	Size        string `json:"size" gorm:"size:50" validate:"required"`
	TotalPrice  int    `json:"total_price" gorm:"not null" validate:"required,gte=0"`
	NmID        int    `json:"nm_id" gorm:"not null" validate:"required"`
	Brand       string `json:"brand" gorm:"size:255" validate:"required"`
	Status      int    `json:"status" gorm:"not null" validate:"required"`
}
