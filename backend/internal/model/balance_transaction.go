package model

import "time"

const (
	BalanceTransactionTypeRecharge = "recharge"
	BalanceTransactionTypePay      = "pay"
	BalanceTransactionTypeRefund   = "refund"
	BalanceTransactionTypeAdjust   = "manual_adjust"

	BalanceTransactionSourceRechargeOrder = "recharge_order"
	BalanceTransactionSourceOrder         = "order"
	BalanceTransactionSourceAdminAdjust   = "admin_adjust"
)

type BalanceTransaction struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	UserID                uint      `gorm:"index" json:"user_id"`
	ChangeAmount          int       `json:"change_amount"`
	PrincipalChangeAmount int       `gorm:"default:0" json:"principal_change_amount"`
	GiftChangeAmount      int       `gorm:"default:0" json:"gift_change_amount"`
	BalanceAfter          int       `json:"balance_after"`
	BizType               string    `gorm:"size:32;index" json:"biz_type"`
	SourceType            string    `gorm:"size:32;index" json:"source_type"`
	SourceID              uint      `gorm:"index" json:"source_id"`
	RelatedBatchID        *uint     `gorm:"index" json:"related_batch_id,omitempty"`
	ShopID                *uint     `gorm:"index" json:"shop_id,omitempty"`
	Remark                string    `gorm:"size:256" json:"remark"`
	CreatedAt             time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}
