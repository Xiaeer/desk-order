package model

type OrderItemOptionValueSnapshot struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceDelta int    `json:"price_delta"`
}

type OrderItemOptionGroupSnapshot struct {
	GroupID    string                         `json:"group_id"`
	GroupName  string                         `json:"group_name"`
	SelectType string                         `json:"select_type"`
	Values     []OrderItemOptionValueSnapshot `json:"values"`
}

type OrderItemOptionGroupSnapshots []OrderItemOptionGroupSnapshot

type OrderItem struct {
	ID              uint                          `gorm:"primaryKey" json:"id"`
	OrderID         uint                          `gorm:"index" json:"order_id"`
	ProductID       uint                          `json:"product_id"`
	Name            string                        `gorm:"size:128" json:"name"`
	Image           string                        `gorm:"size:512" json:"image"`
	Price           int                           `json:"price"` // 单价：分
	SelectedOptions OrderItemOptionGroupSnapshots `gorm:"type:json;serializer:json" json:"selected_options"`
	OptionSummary   string                        `gorm:"size:512" json:"option_summary"`
	Quantity        int                           `json:"quantity"`
	Subtotal        int                           `json:"subtotal"` // 小计：分
}
