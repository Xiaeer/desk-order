package model

import "time"

const (
	UserNotificationCategoryRefund = "refund"
)

type UserNotification struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index" json:"user_id"`
	Category    string     `gorm:"size:32;index" json:"category"`
	Title       string     `gorm:"size:128" json:"title"`
	Content     string     `gorm:"size:512" json:"content"`
	RelatedType string     `gorm:"size:32;index" json:"related_type"`
	RelatedID   uint       `gorm:"index" json:"related_id"`
	IsRead      bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
