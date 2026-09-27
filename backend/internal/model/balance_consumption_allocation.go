package model

import "time"

type BalanceConsumptionAllocation struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index" json:"user_id"`
	OrderID         uint      `gorm:"index:idx_order_batch,unique" json:"order_id"`
	RechargeBatchID uint      `gorm:"index;index:idx_order_batch,unique" json:"recharge_batch_id"`
	AllocationSeq   int       `gorm:"default:0" json:"allocation_seq"`
	TotalAmount     int       `gorm:"default:0" json:"total_amount"`
	PrincipalAmount int       `gorm:"default:0" json:"principal_amount"`
	GiftAmount      int       `gorm:"default:0" json:"gift_amount"`
	CreatedAt       time.Time `json:"created_at"`

	User          User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Order         Order         `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	RechargeBatch RechargeBatch `gorm:"foreignKey:RechargeBatchID" json:"recharge_batch,omitempty"`
}
