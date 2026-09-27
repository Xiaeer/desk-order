package model

import "time"

const (
	ShopAuditStatusNone = -1

	ShopAuditActionSubmit   = "submit"
	ShopAuditActionResubmit = "resubmit"
	ShopAuditActionModify   = "modify"
	ShopAuditActionApprove  = "approve"
	ShopAuditActionReject   = "reject"
	ShopAuditActionDelete   = "delete"

	ShopAuditOperatorMerchant = "merchant"
	ShopAuditOperatorAdmin    = "admin"
)

type ShopAuditRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ShopID       uint      `gorm:"index" json:"shop_id"`
	MerchantID   uint      `gorm:"index" json:"merchant_id"`
	Action       string    `gorm:"size:32;index" json:"action"`
	OperatorRole string    `gorm:"size:16" json:"operator_role"`
	OperatorID   uint      `gorm:"index" json:"operator_id"`
	OperatorName string    `gorm:"size:64" json:"operator_name"`
	FromStatus   int       `json:"from_status"`
	ToStatus     int       `json:"to_status"`
	Remark       string    `gorm:"size:255" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}
