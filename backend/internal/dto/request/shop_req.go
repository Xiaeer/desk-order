package request

// CreateShopReq 创建店铺
type CreateShopReq struct {
	Name        string  `json:"name" binding:"required"`
	Logo        string  `json:"logo"`
	Address     string  `json:"address" binding:"required"`
	Phone       string  `json:"phone" binding:"required"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	Description string  `json:"description"`
}

// UpdateShopReq 更新店铺信息
type UpdateShopReq struct {
	Name        string  `json:"name"`
	Logo        string  `json:"logo"`
	Address     string  `json:"address"`
	Phone       string  `json:"phone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Description string  `json:"description"`
}

// ToggleShopOpenReq 切换营业状态
type ToggleShopOpenReq struct {
	IsOpen bool `json:"is_open"`
}

type ToggleShopAutoAcceptReq struct {
	AutoAcceptOrders bool `json:"auto_accept_orders"`
}

type CreateShopTableReq struct {
	TableNo string `json:"table_no" binding:"required"`
	Sort    int    `json:"sort"`
}

type UpdateShopTableReq struct {
	TableNo *string `json:"table_no"`
	Status  *int    `json:"status"`
	Sort    *int    `json:"sort"`
}
