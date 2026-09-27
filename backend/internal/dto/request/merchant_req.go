package request

// MerchantLoginReq 商家微信登录
type MerchantLoginReq struct {
	Code string `json:"code" binding:"required"`
}

// MerchantRegisterReq 商家注册
type MerchantRegisterReq struct {
	Code  string `json:"code" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

// MerchantWebLoginReq 商户 H5 登录
type MerchantWebLoginReq struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// MerchantWebRegisterReq 商户 H5 注册
type MerchantWebRegisterReq struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// MerchantWebPasswordReq 商户 H5 密码设置/重置
type MerchantWebPasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}
