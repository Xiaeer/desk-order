package model

import "time"

const (
	SubscribeSceneRefundReview = "refund_review"
	SubscribeSceneRefundResult = "refund_result"

	SubscribeGrantStatusPending  = "pending"
	SubscribeGrantStatusConsumed = "consumed"
	SubscribeGrantStatusFailed   = "failed"
)

type UserSubscribeMessageGrant struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index" json:"user_id"`
	Scene        string     `gorm:"size:64;index" json:"scene"`
	TemplateID   string     `gorm:"size:128" json:"template_id"`
	Status       string     `gorm:"size:32;index" json:"status"`
	RelatedType  string     `gorm:"size:32;index" json:"related_type"`
	RelatedID    uint       `gorm:"index" json:"related_id"`
	ErrorMessage string     `gorm:"size:256" json:"error_message"`
	AcceptedAt   time.Time  `json:"accepted_at"`
	ConsumedAt   *time.Time `json:"consumed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
