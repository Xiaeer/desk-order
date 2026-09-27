package model

import "time"

// ShopStatus 店铺审核状态
const (
	ShopStatusPending  = 0 // 待审核
	ShopStatusApproved = 1 // 已通过
	ShopStatusRejected = 2 // 已拒绝
)

type Shop struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	MerchantID       uint      `gorm:"index" json:"merchant_id"`
	Name             string    `gorm:"size:128" json:"name"`
	Logo             string    `gorm:"size:512" json:"logo"`
	Address          string    `gorm:"size:256" json:"address"`
	Phone            string    `gorm:"size:20" json:"phone"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	Status           int       `gorm:"default:0" json:"status"`                 // 0待审核 1通过 2拒绝
	IsOpen           bool      `gorm:"default:false" json:"is_open"`            // 营业状态
	AutoAcceptOrders bool      `gorm:"default:false" json:"auto_accept_orders"` // 自动接单
	PosBindToken     string    `gorm:"size:64" json:"-"`
	Description      string    `gorm:"size:512" json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	Merchant Merchant    `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Tables   []ShopTable `gorm:"foreignKey:ShopID" json:"tables,omitempty"`
}
