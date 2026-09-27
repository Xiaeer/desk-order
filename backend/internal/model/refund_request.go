package model

import "time"

const (
	RefundSourceTypeOrder    = "order"
	RefundSourceTypeRecharge = "recharge"

	RefundRequestChannelUser     = "user"
	RefundRequestChannelAdmin    = "admin"
	RefundRequestChannelMerchant = "merchant"

	RefundReviewScopeOrder    = "order"
	RefundReviewScopeRecharge = "recharge"

	RefundStatusPendingReview = "pending_review"
	RefundStatusApproved      = "approved"
	RefundStatusRejected      = "rejected"
	RefundStatusProcessing    = "processing"
	RefundStatusSuccess       = "success"
	RefundStatusFailed        = "failed"
	RefundStatusCancelled     = "cancelled"
	RefundStatusException     = "exception"
)

type RefundRequest struct {
	ID                       uint       `gorm:"primaryKey" json:"id"`
	RequestNo                string     `gorm:"size:64;uniqueIndex" json:"request_no"`
	SourceType               string     `gorm:"size:32;index" json:"source_type"`
	SourceID                 uint       `gorm:"index" json:"source_id"`
	UserID                   uint       `gorm:"index" json:"user_id"`
	ShopID                   *uint      `gorm:"index" json:"shop_id,omitempty"`
	RequestChannel           string     `gorm:"size:32" json:"request_channel"`
	ReviewScope              string     `gorm:"size:32" json:"review_scope"`
	Status                   string     `gorm:"size:32;index" json:"status"`
	RequestedTotalAmount     int        `gorm:"default:0" json:"requested_total_amount"`
	RequestedPrincipalAmount int        `gorm:"default:0" json:"requested_principal_amount"`
	RequestedGiftVoidAmount  int        `gorm:"default:0" json:"requested_gift_void_amount"`
	ApprovedTotalAmount      int        `gorm:"default:0" json:"approved_total_amount"`
	ApprovedPrincipalAmount  int        `gorm:"default:0" json:"approved_principal_amount"`
	ApprovedGiftVoidAmount   int        `gorm:"default:0" json:"approved_gift_void_amount"`
	RequestNote              string     `gorm:"size:256" json:"request_note"`
	RejectReason             string     `gorm:"size:256" json:"reject_reason"`
	ReviewNote               string     `gorm:"size:256" json:"review_note"`
	ReviewerRole             string     `gorm:"size:32" json:"reviewer_role"`
	ReviewerID               uint       `json:"reviewer_id"`
	ReviewedAt               *time.Time `json:"reviewed_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}
