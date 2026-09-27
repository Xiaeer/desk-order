package model

import "time"

const (
	ProductOptionSelectTypeSingle = "single"
	ProductOptionSelectTypeMulti  = "multi"
)

type ProductOptionValue struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceDelta int    `json:"price_delta"`
	Sort       int    `json:"sort"`
}

type ProductOptionGroup struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	SelectType string               `json:"select_type"`
	Required   bool                 `json:"required"`
	MinSelect  int                  `json:"min_select"`
	MaxSelect  int                  `json:"max_select"`
	Sort       int                  `json:"sort"`
	Values     []ProductOptionValue `json:"values"`
}

type ProductOptionGroups []ProductOptionGroup

type ProductOptionSelection struct {
	GroupID  string   `json:"group_id"`
	ValueIDs []string `json:"value_ids"`
}

type ProductOptionSelections []ProductOptionSelection

type Product struct {
	ID          uint                `gorm:"primaryKey" json:"id"`
	ShopID      uint                `gorm:"index" json:"shop_id"`
	CategoryID  uint                `gorm:"index" json:"category_id"`
	Name        string              `gorm:"size:128" json:"name"`
	Image       string              `gorm:"size:512" json:"image"`
	Price       int                 `json:"price"` // 单位：分
	Description string              `gorm:"size:512" json:"description"`
	Options     ProductOptionGroups `gorm:"type:json;serializer:json" json:"options"`
	IsOnSale    bool                `gorm:"default:true" json:"is_on_sale"`
	Sort        int                 `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`

	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}
