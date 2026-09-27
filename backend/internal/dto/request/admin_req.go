package request

import "time"

// AdminLoginReq 管理员登录
type AdminLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminAuditShopReq 审核店铺
type AdminAuditShopReq struct {
	Status int `json:"status" binding:"required,oneof=1 2"` // 1=通过 2=拒绝
}

// AdminPageReq 分页请求
type AdminPageReq struct {
	Page   int  `form:"page" binding:"omitempty,min=1"`
	Size   int  `form:"size" binding:"omitempty,min=1,max=100"`
	Status *int `form:"status" binding:"omitempty"`
}

// AdminRechargeActivityReq 后台充值活动请求
type AdminRechargeActivityReq struct {
	Name           string     `json:"name" binding:"required"`
	RechargeAmount int        `json:"recharge_amount" binding:"required,min=1"`
	GiftAmount     int        `json:"gift_amount" binding:"omitempty,min=0"`
	Status         *int       `json:"status" binding:"required"`
	Sort           int        `json:"sort"`
	StartAt        *time.Time `json:"start_at"`
	EndAt          *time.Time `json:"end_at"`
	Description    string     `json:"description"`
}

type AdminRechargeOrderListReq struct {
	Page            int    `form:"page" binding:"omitempty,min=1"`
	Size            int    `form:"size" binding:"omitempty,min=1,max=100"`
	Status          *int   `form:"status" binding:"omitempty"`
	OrderNo         string `form:"order_no"`
	UserKeyword     string `form:"user_keyword"`
	SuggestedAction string `form:"suggested_action"`
	CreatedStartAt  string `form:"created_start_at"`
	CreatedEndAt    string `form:"created_end_at"`
}

type AdminRefundRequestListReq struct {
	Page           int    `form:"page" binding:"omitempty,min=1"`
	Size           int    `form:"size" binding:"omitempty,min=1,max=100"`
	Status         string `form:"status"`
	RequestChannel string `form:"request_channel"`
}

type AdminCreateRechargeRefundReq struct {
	RechargeOrderID uint   `json:"recharge_order_id"`
	RechargeOrderNo string `json:"recharge_order_no"`
	RequestNote     string `json:"request_note"`
	ReviewNote      string `json:"review_note"`
}

type AdminReviewRechargeRefundReq struct {
	Action              string `json:"action" binding:"required,oneof=reject complete"`
	ReviewNote          string `json:"review_note"`
	RejectReason        string `json:"reject_reason"`
	ExternalRefundNo    string `json:"external_refund_no"`
	ExternalResponseMsg string `json:"external_response_msg"`
}

type AdminCleanupRechargeOrderReq struct {
	Remark string `json:"remark"`
}

type AdminCleanupUserHistoricalBalanceReq struct {
	Remark string `json:"remark"`
}

// AdminSystemConfigReq 后台系统配置请求
type AdminSystemConfigReq struct {
	MiniUserRefundVisible           *bool  `json:"mini_user_refund_visible" binding:"required"`
	MiniUserRefundApplyEnabled      *bool  `json:"mini_user_refund_apply_enabled" binding:"required"`
	MiniUserBrandName               string `json:"mini_user_brand_name"`
	RefundSubscribePage             string `json:"refund_subscribe_page"`
	RefundReviewSubscribeTemplateID string `json:"refund_review_subscribe_template_id"`
	RefundReviewSubscribeOrderNoKey string `json:"refund_review_subscribe_order_no_key"`
	RefundReviewSubscribeAmountKey  string `json:"refund_review_subscribe_amount_key"`
	RefundReviewSubscribeStatusKey  string `json:"refund_review_subscribe_status_key"`
	RefundReviewSubscribeRemarkKey  string `json:"refund_review_subscribe_remark_key"`
	RefundResultSubscribeTemplateID string `json:"refund_result_subscribe_template_id"`
	RefundResultSubscribeOrderNoKey string `json:"refund_result_subscribe_order_no_key"`
	RefundResultSubscribeAmountKey  string `json:"refund_result_subscribe_amount_key"`
	RefundResultSubscribeStatusKey  string `json:"refund_result_subscribe_status_key"`
	RefundResultSubscribeRemarkKey  string `json:"refund_result_subscribe_remark_key"`
}
