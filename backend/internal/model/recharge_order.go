package model

import "time"

const (
	RechargeOrderStatusPending   = 0
	RechargeOrderStatusPaid      = 1
	RechargeOrderStatusCancelled = 2
)

type RechargeOrder struct {
	ID                      uint       `gorm:"primaryKey" json:"id"`
	OrderNo                 string     `gorm:"uniqueIndex;size:64" json:"order_no"`
	UserID                  uint       `gorm:"index" json:"user_id"`
	ActivityID              uint       `gorm:"index" json:"activity_id"`
	ActivityNameSnapshot    string     `gorm:"size:128" json:"activity_name_snapshot"`
	ActivityVersionSnapshot int        `gorm:"default:1" json:"activity_version_snapshot"`
	RechargeAmount          int        `json:"recharge_amount"`
	GiftAmount              int        `gorm:"default:0" json:"gift_amount"`
	PrincipalConsumeWeight  int        `gorm:"default:0" json:"principal_consume_weight"`
	GiftConsumeWeight       int        `gorm:"default:0" json:"gift_consume_weight"`
	TotalArrivalAmount      int        `json:"total_arrival_amount"`
	PayAmount               int        `json:"pay_amount"`
	PayChannel              string     `gorm:"size:32;default:'wechat'" json:"pay_channel"`
	Status                  int        `gorm:"default:0;index" json:"status"`
	ExpiresAt               *time.Time `json:"expires_at"`
	PaidAt                  *time.Time `json:"paid_at"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`

	User     User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Activity RechargeActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
}
