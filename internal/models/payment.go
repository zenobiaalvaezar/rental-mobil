package models

import "time"

type Payment struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	RentalID   uint          `gorm:"not null" json:"rental_id"`
	InvoiceID  string        `gorm:"not null" json:"invoice_id"`
	Amount     float64       `gorm:"not null" json:"amount"`
	Status     string        `gorm:"not null" json:"status"`
	PaymentURL string        `gorm:"not null" json:"payment_url"`
	ExternalID string        `gorm:"not null" json:"external_id"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Rental     RentalHistory `gorm:"foreignKey:RentalID" json:"rental"`
}
