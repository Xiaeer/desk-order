package response

import "time"

// UserLoginResp 用户登录响应
type UserLoginResp struct {
	Token string `json:"token"`
}

// NearbyShopResp 附近商家项
type NearbyShopResp struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Logo        string  `json:"logo"`
	Address     string  `json:"address"`
	Distance    float64 `json:"distance"` // 单位：km
	IsOpen      bool    `json:"is_open"`
	Description string  `json:"description"`
}

// MiniProgramPayResp 小程序微信支付参数
type MiniProgramPayResp struct {
	TimeStamp  string `json:"timeStamp,omitempty"`
	NonceStr   string `json:"nonceStr,omitempty"`
	Package    string `json:"package,omitempty"`
	SignType   string `json:"signType,omitempty"`
	PaySign    string `json:"paySign,omitempty"`
	Mode       string `json:"mode,omitempty"`
	MockResult string `json:"mock_result,omitempty"`
}

// UserProfileResp 用户资料与钱包信息
type UserProfileResp struct {
	ID                         uint     `json:"id"`
	Nickname                   string   `json:"nickname"`
	Avatar                     string   `json:"avatar"`
	Phone                      string   `json:"phone"`
	BalanceAmount              int      `json:"balance_amount"`
	TotalRechargeAmount        int      `json:"total_recharge_amount"`
	TotalGiftAmount            int      `json:"total_gift_amount"`
	TotalConsumeAmount         int      `json:"total_consume_amount"`
	RefundUnreadCount          int64    `json:"refund_unread_count"`
	HasRefundRecords           bool     `json:"has_refund_records"`
	RefundSubscribeEnabled     bool     `json:"refund_subscribe_enabled"`
	RefundSubscribeTemplateIDs []string `json:"refund_subscribe_template_ids"`
	RefundFeatureVisible       bool     `json:"refund_feature_visible"`
	RefundFeatureEnabled       bool     `json:"refund_feature_enabled"`
}

// RechargeActivityResp 充值活动响应
type RechargeActivityResp struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	RechargeAmount int        `json:"recharge_amount"`
	GiftAmount     int        `json:"gift_amount"`
	ArrivalAmount  int        `json:"arrival_amount"`
	Status         int        `json:"status"`
	Sort           int        `json:"sort"`
	StartAt        *time.Time `json:"start_at"`
	EndAt          *time.Time `json:"end_at"`
	Description    string     `json:"description"`
}

// RechargeOrderResp 充值订单响应
type RechargeOrderResp struct {
	ID                 uint       `json:"id"`
	OrderNo            string     `json:"order_no"`
	ActivityID         uint       `json:"activity_id"`
	ActivityName       string     `json:"activity_name"`
	RechargeAmount     int        `json:"recharge_amount"`
	GiftAmount         int        `json:"gift_amount"`
	TotalArrivalAmount int        `json:"total_arrival_amount"`
	PayAmount          int        `json:"pay_amount"`
	PayChannel         string     `json:"pay_channel"`
	Status             int        `json:"status"`
	CanApplyRefund     bool       `json:"can_apply_refund"`
	RefundEntryVisible bool       `json:"refund_entry_visible"`
	RefundActionText   string     `json:"refund_action_text"`
	RefundStatus       string     `json:"refund_status"`
	RefundStatusText   string     `json:"refund_status_text"`
	RefundResultText   string     `json:"refund_result_text"`
	RefundUpdatedAt    *time.Time `json:"refund_updated_at"`
	PaidAt             *time.Time `json:"paid_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

// BalanceTransactionResp 余额流水响应
type BalanceTransactionResp struct {
	ID           uint      `json:"id"`
	ChangeAmount int       `json:"change_amount"`
	BalanceAfter int       `json:"balance_after"`
	BizType      string    `json:"biz_type"`
	SourceType   string    `json:"source_type"`
	SourceID     uint      `json:"source_id"`
	ShopID       *uint     `json:"shop_id,omitempty"`
	ShopName     string    `json:"shop_name,omitempty"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserRefundRecordResp 用户退款记录响应
type UserRefundRecordResp struct {
	ID                   uint      `json:"id"`
	RequestNo            string    `json:"request_no"`
	SourceType           string    `json:"source_type"`
	SourceID             uint      `json:"source_id"`
	SourceNo             string    `json:"source_no"`
	SourceTitle          string    `json:"source_title"`
	RequestNote          string    `json:"request_note"`
	CanCancel            bool      `json:"can_cancel"`
	CancelActionText     string    `json:"cancel_action_text"`
	Status               string    `json:"status"`
	StatusText           string    `json:"status_text"`
	RequestedTotalAmount int       `json:"requested_total_amount"`
	ApprovedTotalAmount  int       `json:"approved_total_amount"`
	ResultText           string    `json:"result_text"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// UserNotificationResp 用户通知响应
type UserNotificationResp struct {
	ID          uint      `json:"id"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	RelatedType string    `json:"related_type"`
	RelatedID   uint      `json:"related_id"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserPublicConfigResp struct {
	MiniUserBrandName string `json:"mini_user_brand_name"`
}
