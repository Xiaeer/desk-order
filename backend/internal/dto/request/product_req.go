package request

import "deskorder/internal/model"

// CreateCategoryReq 创建分类
type CreateCategoryReq struct {
	Name string `json:"name" binding:"required"`
	Sort int    `json:"sort"`
}

// UpdateCategoryReq 更新分类
type UpdateCategoryReq struct {
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

// CreateProductReq 创建菜品
type CreateProductReq struct {
	CategoryID  uint                      `json:"category_id" binding:"required"`
	Name        string                    `json:"name" binding:"required"`
	Image       string                    `json:"image"`
	Price       int                       `json:"price" binding:"required,min=1"` // 单位：分
	Description string                    `json:"description"`
	Options     model.ProductOptionGroups `json:"options"`
	Sort        int                       `json:"sort"`
}

// UpdateProductReq 更新菜品
type UpdateProductReq struct {
	CategoryID  uint                       `json:"category_id"`
	Name        string                     `json:"name"`
	Image       string                     `json:"image"`
	Price       int                        `json:"price"`
	Description string                     `json:"description"`
	Options     *model.ProductOptionGroups `json:"options"`
	IsOnSale    *bool                      `json:"is_on_sale"` // 指针区分零值
	Sort        int                        `json:"sort"`
}
