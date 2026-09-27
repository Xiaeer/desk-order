package response

type AdminUserResp struct {
	ID                                 uint                     `json:"id"`
	Nickname                           string                   `json:"nickname"`
	DisplayName                        string                   `json:"display_name"`
	IdentityLabel                      string                   `json:"identity_label"`
	Avatar                             string                   `json:"avatar"`
	Phone                              string                   `json:"phone"`
	BalanceAmount                      int                      `json:"balance_amount"`
	TotalRechargeAmount                int                      `json:"total_recharge_amount"`
	TotalGiftAmount                    int                      `json:"total_gift_amount"`
	TotalConsumeAmount                 int                      `json:"total_consume_amount"`
	HistoricalBalanceBatchID           uint                     `json:"historical_balance_batch_id"`
	HistoricalBalanceAmount            int                      `json:"historical_balance_amount"`
	CanCleanupHistoricalBalance        bool                     `json:"can_cleanup_historical_balance"`
	HistoricalBalanceUnavailableReason string                   `json:"historical_balance_unavailable_reason"`
	OrderCount                         int64                    `json:"order_count"`
	RechargeOrderCount                 int64                    `json:"recharge_order_count"`
	RecentRechargeOrders               []AdminRechargeOrderResp `json:"recent_recharge_orders,omitempty"`
	CreatedAt                          string                   `json:"created_at"`
	UpdatedAt                          string                   `json:"updated_at"`
}

type AdminLoginResp struct {
	Token    string `json:"token"`
	Nickname string `json:"nickname"`
}

type DashboardResp struct {
	MerchantCount int64 `json:"merchant_count"`
	ShopCount     int64 `json:"shop_count"`
	OrderCount    int64 `json:"order_count"`
	UserCount     int64 `json:"user_count"`
	TotalAmount   int64 `json:"total_amount"`
	PendingShops  int64 `json:"pending_shops"`
}

type AdminMerchantResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
	HasShop   bool   `json:"has_shop"`
	ShopName  string `json:"shop_name"`
}

type AdminShopResp struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Logo         string  `json:"logo"`
	Address      string  `json:"address"`
	Phone        string  `json:"phone"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Status       int     `json:"status"`
	IsOpen       bool    `json:"is_open"`
	Description  string  `json:"description"`
	MerchantID   uint    `json:"merchant_id"`
	MerchantName string  `json:"merchant_name"`
	CreatedAt    string  `json:"created_at"`
}

type ShopAuditRecordResp struct {
	ID           uint   `json:"id"`
	ShopID       uint   `json:"shop_id"`
	MerchantID   uint   `json:"merchant_id"`
	Action       string `json:"action"`
	ActionText   string `json:"action_text"`
	OperatorRole string `json:"operator_role"`
	OperatorID   uint   `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	FromStatus   int    `json:"from_status"`
	ToStatus     int    `json:"to_status"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"created_at"`
}

type AdminRechargeActivityResp struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	RechargeAmount int    `json:"recharge_amount"`
	GiftAmount     int    `json:"gift_amount"`
	ArrivalAmount  int    `json:"arrival_amount"`
	Status         int    `json:"status"`
	Sort           int    `json:"sort"`
	StartAt        string `json:"start_at"`
	EndAt          string `json:"end_at"`
	Description    string `json:"description"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type AdminRechargeOrderResp struct {
	ID                        uint   `json:"id"`
	OrderNo                   string `json:"order_no"`
	UserID                    uint   `json:"user_id"`
	UserDisplayName           string `json:"user_display_name"`
	UserIdentityLabel         string `json:"user_identity_label"`
	UserNickname              string `json:"user_nickname"`
	UserPhone                 string `json:"user_phone"`
	ActivityID                uint   `json:"activity_id"`
	ActivityName              string `json:"activity_name"`
	ActivityVersion           int    `json:"activity_version"`
	RechargeAmount            int    `json:"recharge_amount"`
	GiftAmount                int    `json:"gift_amount"`
	TotalArrivalAmount        int    `json:"total_arrival_amount"`
	PayAmount                 int    `json:"pay_amount"`
	PayChannel                string `json:"pay_channel"`
	PayChannelText            string `json:"pay_channel_text"`
	Status                    int    `json:"status"`
	StatusText                string `json:"status_text"`
	ExpiresAt                 string `json:"expires_at"`
	PaidAt                    string `json:"paid_at"`
	CreatedAt                 string `json:"created_at"`
	UpdatedAt                 string `json:"updated_at"`
	RefundRequestID           uint   `json:"refund_request_id"`
	RefundRequestNo           string `json:"refund_request_no"`
	RefundStatus              string `json:"refund_status"`
	RefundStatusText          string `json:"refund_status_text"`
	RefundRequestedAmount     int    `json:"refund_requested_amount"`
	RefundApprovedAmount      int    `json:"refund_approved_amount"`
	RefundRequestChannel      string `json:"refund_request_channel"`
	RefundRequestChannelText  string `json:"refund_request_channel_text"`
	RefundUpdatedAt           string `json:"refund_updated_at"`
	BatchStatus               string `json:"batch_status"`
	BatchStatusText           string `json:"batch_status_text"`
	SuggestedAction           string `json:"suggested_action"`
	SuggestedActionText       string `json:"suggested_action_text"`
	CanCreateRefund           bool   `json:"can_create_refund"`
	RefundUnavailableReason   string `json:"refund_unavailable_reason"`
	CurrentPrincipalRemaining int    `json:"current_principal_remaining"`
	CurrentGiftRemaining      int    `json:"current_gift_remaining"`
	CurrentRemainingAmount    int    `json:"current_remaining_amount"`
	HasRemainingSnapshot      bool   `json:"has_remaining_snapshot"`
	CanCleanupRemaining       bool   `json:"can_cleanup_remaining"`
	CleanupUnavailableReason  string `json:"cleanup_unavailable_reason"`
}

type AdminRefundRequestResp struct {
	ID                        uint   `json:"id"`
	RequestNo                 string `json:"request_no"`
	SourceType                string `json:"source_type"`
	ReviewScope               string `json:"review_scope"`
	RequestChannel            string `json:"request_channel"`
	RequestChannelText        string `json:"request_channel_text"`
	SourceID                  uint   `json:"source_id"`
	RechargeOrderID           uint   `json:"recharge_order_id"`
	RechargeOrderNo           string `json:"recharge_order_no"`
	UserID                    uint   `json:"user_id"`
	UserNickname              string `json:"user_nickname"`
	UserPhone                 string `json:"user_phone"`
	Status                    string `json:"status"`
	RequestedTotalAmount      int    `json:"requested_total_amount"`
	RequestedPrincipalAmount  int    `json:"requested_principal_amount"`
	RequestedGiftVoidAmount   int    `json:"requested_gift_void_amount"`
	ApprovedTotalAmount       int    `json:"approved_total_amount"`
	ApprovedPrincipalAmount   int    `json:"approved_principal_amount"`
	ApprovedGiftVoidAmount    int    `json:"approved_gift_void_amount"`
	CurrentPrincipalRemaining int    `json:"current_principal_remaining"`
	CurrentGiftRemaining      int    `json:"current_gift_remaining"`
	CurrentBalanceAmount      int    `json:"current_balance_amount"`
	RequestNote               string `json:"request_note"`
	ReviewNote                string `json:"review_note"`
	RejectReason              string `json:"reject_reason"`
	ExecutionStatus           string `json:"execution_status"`
	SyncRetryCount            int    `json:"sync_retry_count"`
	LastSyncAt                string `json:"last_sync_at"`
	NextAutoSyncAt            string `json:"next_auto_sync_at"`
	AutoSyncPaused            bool   `json:"auto_sync_paused"`
	ExternalRefundNo          string `json:"external_refund_no"`
	ExternalResponseCode      string `json:"external_response_code"`
	ExternalResponseMsg       string `json:"external_response_msg"`
	CreatedAt                 string `json:"created_at"`
	ReviewedAt                string `json:"reviewed_at"`
}

type AdminSystemConfigResp struct {
	MiniUserRefundVisible           bool   `json:"mini_user_refund_visible"`
	MiniUserRefundApplyEnabled      bool   `json:"mini_user_refund_apply_enabled"`
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
