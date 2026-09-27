package request

import "deskorder/internal/model"

// UserLoginReq 用户微信登录请求
type UserLoginReq struct {
	Code string `json:"code" binding:"required"`
}

// UpdateUserProfileReq 更新用户资料请求
type UpdateUserProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// BindUserPhoneReq 绑定用户手机号请求
type BindUserPhoneReq struct {
	Code string `json:"code" binding:"required"`
}

// NearbyShopsReq 获取附近商家请求
type NearbyShopsReq struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
}

// CreateOrderReq 用户下单请求
type CreateOrderReq struct {
	ShopID      uint           `json:"shop_id" binding:"required"`
	ShopTableID *uint          `json:"shop_table_id"`
	EntryScene  string         `json:"entry_scene"`
	Remark      string         `json:"remark"`
	Items       []OrderItemReq `json:"items" binding:"required,min=1"`
}

// OrderItemReq 下单商品项
type OrderItemReq struct {
	ProductID       uint                          `json:"product_id" binding:"required"`
	Quantity        int                           `json:"quantity" binding:"required,min=1"`
	SelectedOptions model.ProductOptionSelections `json:"selected_options"`
}

// UserOrderListReq 用户订单列表请求
type UserOrderListReq struct {
	Page int `form:"page" binding:"required,min=1"`
	Size int `form:"size" binding:"required,min=1,max=50"`
}

// OrderPayReq 用户订单支付请求
type OrderPayReq struct {
	Method string `json:"method"`
}

// CreateRechargeOrderReq 创建充值订单请求
type CreateRechargeOrderReq struct {
	ActivityID uint `json:"activity_id" binding:"required"`
}

// RechargeOrderListReq 充值订单列表请求
type RechargeOrderListReq struct {
	Page   int  `form:"page" binding:"required,min=1"`
	Size   int  `form:"size" binding:"required,min=1,max=50"`
	Status *int `form:"status"`
}

// ApplyRechargeRefundReq 用户申请充值退款
type ApplyRechargeRefundReq struct {
	RequestNote string `json:"request_note"`
}

// RefundSubscribeGrantReq 用户上报订阅消息授权结果
type RefundSubscribeGrantReq struct {
	TemplateIDs []string `json:"template_ids"`
}

// CancelRefundRequestReq 用户撤回退款申请
type CancelRefundRequestReq struct{}

// BalanceTransactionListReq 余额流水列表请求
type BalanceTransactionListReq struct {
	Page int `form:"page" binding:"required,min=1"`
	Size int `form:"size" binding:"required,min=1,max=50"`
}

// UserPageReq 通用分页请求
type UserPageReq struct {
	Page int `form:"page" binding:"required,min=1"`
	Size int `form:"size" binding:"required,min=1,max=50"`
}
