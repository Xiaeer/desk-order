package request

import "deskorder/internal/model"

type POSAuthReq struct {
	ShopID   uint   `json:"shop_id" binding:"required"`
	POSToken string `json:"pos_token" binding:"required"`
}

type POSCreateOrderReq struct {
	ShopID          uint           `json:"shop_id" binding:"required"`
	POSToken        string         `json:"pos_token" binding:"required"`
	ClientRequestID string         `json:"client_request_id" binding:"required"`
	TableNo         string         `json:"table_no"`
	Remark          string         `json:"remark"`
	Items           []OrderItemReq `json:"items" binding:"required,min=1"`
	TerminalRef     string         `json:"terminal_ref"`
}

type POSOrderPayReq struct {
	ShopID   uint   `json:"shop_id" binding:"required"`
	POSToken string `json:"pos_token" binding:"required"`
}

type POSOrderCancelReq struct {
	ShopID   uint   `json:"shop_id" binding:"required"`
	POSToken string `json:"pos_token" binding:"required"`
}

type POSOrderStatusReq struct {
	ShopID   uint   `form:"shop_id" binding:"required"`
	POSToken string `form:"pos_token" binding:"required"`
}

type POSOrderItemReq struct {
	ProductID       uint                          `json:"product_id" binding:"required"`
	Quantity        int                           `json:"quantity" binding:"required,min=1"`
	SelectedOptions model.ProductOptionSelections `json:"selected_options"`
}
