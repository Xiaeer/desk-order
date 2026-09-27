package handler

import (
	"deskorder/internal/dto/request"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

// MerchantLogin 商家微信登录
func MerchantLogin(c *gin.Context) {
	var req request.MerchantLoginReq
	if !utils.BindJSON(c, &req) {
		return
	}

	resp, err := service.MerchantLogin(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// MerchantRegister 商家注册
func MerchantRegister(c *gin.Context) {
	var req request.MerchantRegisterReq
	if !utils.BindJSON(c, &req) {
		return
	}

	resp, err := service.MerchantRegister(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// MerchantWebRegister 商户 H5 注册
func MerchantWebRegister(c *gin.Context) {
	var req request.MerchantWebRegisterReq
	if !utils.BindJSON(c, &req) {
		return
	}

	resp, err := service.MerchantWebRegister(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// MerchantWebLogin 商户 H5 登录
func MerchantWebLogin(c *gin.Context) {
	var req request.MerchantWebLoginReq
	if !utils.BindJSON(c, &req) {
		return
	}

	resp, err := service.MerchantWebLogin(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// MerchantUpdateWebPassword 商户设置/重置 H5 登录密码
func MerchantUpdateWebPassword(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.MerchantWebPasswordReq
	if !utils.BindJSON(c, &req) {
		return
	}

	if err := service.MerchantUpdateWebPassword(merchantID, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// MerchantResetWebPasswordByMiniProgram 商户在小程序内设置/重置 H5 登录密码
func MerchantResetWebPasswordByMiniProgram(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.MerchantWebPasswordReq
	if !utils.BindJSON(c, &req) {
		return
	}

	if err := service.MerchantResetWebPasswordByMiniProgram(merchantID, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
