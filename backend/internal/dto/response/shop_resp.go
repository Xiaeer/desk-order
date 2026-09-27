package response

type ShopResp struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Logo        string  `json:"logo"`
	Address     string  `json:"address"`
	Phone       string  `json:"phone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Status      int     `json:"status"`
	IsOpen      bool    `json:"is_open"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance,omitempty"` // 距离(km)，仅附近商家列表返回
}

type MerchantShopResp struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Logo             string  `json:"logo"`
	Address          string  `json:"address"`
	Phone            string  `json:"phone"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Status           int     `json:"status"`
	IsOpen           bool    `json:"is_open"`
	AutoAcceptOrders bool    `json:"auto_accept_orders"`
	Description      string  `json:"description"`
	PosBindToken     string  `json:"pos_bind_token"`
}

type ShopTableResp struct {
	ID         uint   `json:"id"`
	ShopID     uint   `json:"shop_id"`
	TableNo    string `json:"table_no"`
	SceneToken string `json:"scene_token"`
	QRCodeURL  string `json:"qrcode_url"`
	Status     int    `json:"status"`
	Sort       int    `json:"sort"`
}

type ShopTableSceneResp struct {
	ShopID      uint   `json:"shop_id"`
	ShopName    string `json:"shop_name"`
	ShopTableID uint   `json:"shop_table_id"`
	TableNo     string `json:"table_no"`
	SceneToken  string `json:"scene_token"`
}
