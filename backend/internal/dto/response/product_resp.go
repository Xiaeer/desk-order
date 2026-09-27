package response

import "deskorder/internal/model"

type CategoryResp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

type ProductResp struct {
	ID          uint                      `json:"id"`
	CategoryID  uint                      `json:"category_id"`
	Name        string                    `json:"name"`
	Image       string                    `json:"image"`
	Price       int                       `json:"price"`
	Description string                    `json:"description"`
	Options     model.ProductOptionGroups `json:"options"`
	IsOnSale    bool                      `json:"is_on_sale"`
	Sort        int                       `json:"sort"`
}

// MenuResp 菜单响应（分类+菜品）
type MenuResp struct {
	ShopID     uint                   `json:"shop_id"`
	ShopName   string                 `json:"shop_name"`
	Categories []CategoryWithProducts `json:"categories"`
}

type CategoryWithProducts struct {
	CategoryResp
	Products []ProductResp `json:"products"`
}
