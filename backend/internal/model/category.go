package model

import "time"

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ShopID    uint      `gorm:"index" json:"shop_id"`
	Name      string    `gorm:"size:64" json:"name"`
	Sort      int       `gorm:"default:0" json:"sort"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
