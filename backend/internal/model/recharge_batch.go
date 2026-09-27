package model

import "time"

const (
	RechargeBatchTypeRecharge = "recharge"
	RechargeBatchTypeOpening  = "opening"
)

type RechargeBatch struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	UserID                  uint      `gorm:"index" json:"user_id"`
	RechargeOrderID         *uint     `gorm:"uniqueIndex" json:"recharge_order_id,omitempty"`
	ActivityID              *uint     `gorm:"index" json:"activity_id,omitempty"`
	ActivityNameSnapshot    string    `gorm:"size:128" json:"activity_name_snapshot"`
	ActivityVersionSnapshot int       `gorm:"default:1" json:"activity_version_snapshot"`
	BatchType               string    `gorm:"size:32;index;default:'recharge'" json:"batch_type"`
	PrincipalTotal          int       `gorm:"default:0" json:"principal_total"`
	GiftTotal               int       `gorm:"default:0" json:"gift_total"`
	PrincipalRemaining      int       `gorm:"default:0" json:"principal_remaining"`
	GiftRemaining           int       `gorm:"default:0" json:"gift_remaining"`
	PrincipalConsumeWeight  int       `gorm:"default:0" json:"principal_consume_weight"`
	GiftConsumeWeight       int       `gorm:"default:0" json:"gift_consume_weight"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`

	User          User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RechargeOrder *RechargeOrder    `gorm:"foreignKey:RechargeOrderID" json:"recharge_order,omitempty"`
	Activity      *RechargeActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
}
