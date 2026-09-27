package model

import "time"

const (
	ShopTableStatusEnabled  = 0
	ShopTableStatusDisabled = 1
)

type ShopTable struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ShopID     uint      `gorm:"index;uniqueIndex:idx_shop_table_no" json:"shop_id"`
	TableNo    string    `gorm:"size:32;index:idx_shop_table_no,unique" json:"table_no"`
	SceneToken string    `gorm:"size:64;uniqueIndex" json:"scene_token"`
	QRCodeURL  string    `gorm:"size:512" json:"qrcode_url"`
	Status     int       `gorm:"default:0" json:"status"`
	Sort       int       `gorm:"default:0" json:"sort"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}
