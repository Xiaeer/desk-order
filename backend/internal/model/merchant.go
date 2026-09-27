package model

import "time"

type Merchant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"uniqueIndex;size:64" json:"-"`
	Phone     string    `gorm:"size:20" json:"phone"`
	Password  string    `gorm:"size:128" json:"-"`
	Name      string    `gorm:"size:64" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
