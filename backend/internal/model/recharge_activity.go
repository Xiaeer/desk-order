package model

import "time"

const (
	RechargeActivityStatusDisabled = 0
	RechargeActivityStatusEnabled  = 1
)

type RechargeActivity struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:128" json:"name"`
	RechargeAmount int        `json:"recharge_amount"`
	GiftAmount     int        `gorm:"default:0" json:"gift_amount"`
	RuleVersion    int        `gorm:"default:1" json:"rule_version"`
	Status         int        `gorm:"default:1;index" json:"status"`
	Sort           int        `gorm:"default:0;index" json:"sort"`
	StartAt        *time.Time `json:"start_at"`
	EndAt          *time.Time `json:"end_at"`
	Description    string     `gorm:"size:256" json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
