package model

import "time"

// OrderStatus 订单状态
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付
	OrderStatusAccepted  = 2 // 已接单
	OrderStatusCompleted = 3 // 已完成
	OrderStatusCancelled = 4 // 已取消

	PaymentChannelWechat  = "wechat"
	PaymentChannelBalance = "balance"
)

type Order struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	OrderNo           string     `gorm:"uniqueIndex;size:64" json:"order_no"`
	UserID            uint       `gorm:"index" json:"user_id"`
	ShopID            uint       `gorm:"index;index:idx_order_shop_client_request_id,unique" json:"shop_id"`
	ClientRequestID   *string    `gorm:"size:64;index:idx_order_shop_client_request_id,unique" json:"client_request_id,omitempty"`
	ShopTableID       *uint      `gorm:"index" json:"shop_table_id"`
	TableNoSnapshot   string     `gorm:"size:32" json:"table_no_snapshot"`
	EntryScene        string     `gorm:"size:64" json:"entry_scene"`
	TotalAmount       int        `json:"total_amount"` // 单位：分
	PayChannel        string     `gorm:"size:32;default:'wechat'" json:"pay_channel"`
	BalancePaidAmount int        `gorm:"default:0" json:"balance_paid_amount"`
	WechatPaidAmount  int        `gorm:"default:0" json:"wechat_paid_amount"`
	PaymentOrderNo    string     `gorm:"size:64" json:"payment_order_no"`
	Status            int        `gorm:"default:0" json:"status"`
	Remark            string     `gorm:"size:256" json:"remark"`
	PaidAt            *time.Time `json:"paid_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Shop      Shop        `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	ShopTable ShopTable   `gorm:"foreignKey:ShopTableID" json:"shop_table,omitempty"`
}
