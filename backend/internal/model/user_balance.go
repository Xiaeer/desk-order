package model

import "time"

type UserBalance struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	UserID              uint      `gorm:"uniqueIndex" json:"user_id"`
	BalanceAmount       int       `gorm:"default:0" json:"balance_amount"`
	TotalRechargeAmount int       `gorm:"default:0" json:"total_recharge_amount"`
	TotalGiftAmount     int       `gorm:"default:0" json:"total_gift_amount"`
	TotalConsumeAmount  int       `gorm:"default:0" json:"total_consume_amount"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
