package response

import (
	"time"

	"deskorder/internal/model"
)

type OrderResp struct {
	ID                 uint            `json:"id"`
	OrderNo            string          `json:"order_no"`
	UserID             uint            `json:"user_id"`
	ShopID             uint            `json:"shop_id"`
	ShopTableID        *uint           `json:"shop_table_id"`
	TableNoSnapshot    string          `json:"table_no_snapshot"`
	EntryScene         string          `json:"entry_scene"`
	ShopName           string          `json:"shop_name"`
	TotalAmount        int             `json:"total_amount"`
	PayChannel         string          `json:"pay_channel"`
	BalancePaidAmount  int             `json:"balance_paid_amount"`
	WechatPaidAmount   int             `json:"wechat_paid_amount"`
	Status             int             `json:"status"`
	Remark             string          `json:"remark"`
	PaidAt             *time.Time      `json:"paid_at"`
	CreatedAt          time.Time       `json:"created_at"`
	CanUserCancel      bool            `json:"can_user_cancel"`
	CanUserRefund      bool            `json:"can_user_refund"`
	RefundEntryVisible bool            `json:"refund_entry_visible"`
	RefundActionText   string          `json:"refund_action_text"`
	Items              []OrderItemResp `json:"items,omitempty"`
}

type OrderItemResp struct {
	ID              uint                                `json:"id"`
	ProductID       uint                                `json:"product_id"`
	Name            string                              `json:"name"`
	Image           string                              `json:"image"`
	Price           int                                 `json:"price"`
	SelectedOptions model.OrderItemOptionGroupSnapshots `json:"selected_options"`
	OptionSummary   string                              `json:"option_summary"`
	Quantity        int                                 `json:"quantity"`
	Subtotal        int                                 `json:"subtotal"`
}
