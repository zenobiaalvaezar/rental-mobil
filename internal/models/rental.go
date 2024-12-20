package models

import "time"

type RentalHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	CarID       uint      `gorm:"not null" json:"car_id"`
	RentalStart time.Time `gorm:"not null" json:"rental_start"`
	RentalEnd   time.Time `gorm:"not null" json:"rental_end"`
	TotalCost   float64   `gorm:"not null" json:"total_cost"`
	Status      string    `gorm:"not null" json:"status"` // pending/active/completed/cancelled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	Car         Car       `gorm:"foreignKey:CarID" json:"car"`
}

func (RentalHistory) TableName() string {
	return "rental_history"
}
