package model

import "time"

const (
	SystemConfigKeyMiniUserRefundVisible           = "mini_user_refund_visible"
	SystemConfigKeyMiniUserRefundApplyEnabled      = "mini_user_refund_apply_enabled"
	SystemConfigKeyMiniUserBrandName               = "mini_user_brand_name"
	SystemConfigValueTypeBool                      = "bool"
	SystemConfigValueTypeString                    = "string"
	SystemConfigKeyRefundSubscribePage             = "refund_subscribe_page"
	SystemConfigKeyRefundReviewSubscribeTemplateID = "refund_review_subscribe_template_id"
	SystemConfigKeyRefundReviewSubscribeOrderNoKey = "refund_review_subscribe_order_no_key"
	SystemConfigKeyRefundReviewSubscribeAmountKey  = "refund_review_subscribe_amount_key"
	SystemConfigKeyRefundReviewSubscribeStatusKey  = "refund_review_subscribe_status_key"
	SystemConfigKeyRefundReviewSubscribeRemarkKey  = "refund_review_subscribe_remark_key"
	SystemConfigKeyRefundResultSubscribeTemplateID = "refund_result_subscribe_template_id"
	SystemConfigKeyRefundResultSubscribeOrderNoKey = "refund_result_subscribe_order_no_key"
	SystemConfigKeyRefundResultSubscribeAmountKey  = "refund_result_subscribe_amount_key"
	SystemConfigKeyRefundResultSubscribeStatusKey  = "refund_result_subscribe_status_key"
	SystemConfigKeyRefundResultSubscribeRemarkKey  = "refund_result_subscribe_remark_key"
)

type SystemConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ConfigKey   string    `gorm:"size:64;uniqueIndex" json:"config_key"`
	ConfigValue string    `gorm:"size:256" json:"config_value"`
	ValueType   string    `gorm:"size:32" json:"value_type"`
	Description string    `gorm:"size:256" json:"description"`
	UpdatedBy   uint      `json:"updated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
