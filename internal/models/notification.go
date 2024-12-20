package models

import "time"

type UserNotification struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	Type        string    `gorm:"not null" json:"type"`         // registration/login/topup/rental/return
	EmailStatus string    `gorm:"not null" json:"email_status"` // sent/failed
	Message     string    `gorm:"not null" json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
}
