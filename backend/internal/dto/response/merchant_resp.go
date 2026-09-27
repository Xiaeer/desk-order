package response

type MerchantLoginResp struct {
	Token   string `json:"token"`
	IsNew   bool   `json:"is_new"`
	HasShop bool   `json:"has_shop"`
}
