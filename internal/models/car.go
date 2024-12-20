package models

import "time"

type Car struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"not null" json:"name"`
	StockAvailability int       `gorm:"not null" json:"stock_availability"`
	RentalCosts       float64   `gorm:"not null" json:"rental_costs"`
	Category          string    `gorm:"not null" json:"category"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
