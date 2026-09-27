package handler

import (
	"strconv"
	"strings"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

func toMerchantShopResp(shop *model.Shop) response.MerchantShopResp {
	if shop == nil {
		return response.MerchantShopResp{}
	}
	return response.MerchantShopResp{
		ID:               shop.ID,
		Name:             shop.Name,
		Logo:             shop.Logo,
		Address:          shop.Address,
		Phone:            shop.Phone,
		Latitude:         shop.Latitude,
		Longitude:        shop.Longitude,
		Status:           shop.Status,
		IsOpen:           shop.IsOpen,
		AutoAcceptOrders: shop.AutoAcceptOrders,
		Description:      shop.Description,
		PosBindToken:     shop.PosBindToken,
	}
}

// CreateShop 创建店铺
func CreateShop(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.CreateShopReq
	if !utils.BindJSON(c, &req) {
		return
	}

	shop, err := service.CreateShop(merchantID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, toMerchantShopResp(shop))
}

// GetMerchantShop 获取商家店铺信息
func GetMerchantShop(c *gin.Context) {
	merchantID := c.GetUint("userID")
	shop, err := service.GetMerchantShop(merchantID)
	if err != nil {
		utils.Fail(c, 404, "店铺不存在")
		return
	}
	utils.Success(c, toMerchantShopResp(shop))
}

// UpdateShop 更新店铺信息
func UpdateShop(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.UpdateShopReq
	if !utils.BindJSON(c, &req) {
		return
	}

	shop, err := service.UpdateShop(merchantID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, toMerchantShopResp(shop))
}

// ToggleShopOpen 切换营业状态
func ToggleShopOpen(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.ToggleShopOpenReq
	if !utils.BindJSON(c, &req) {
		return
	}

	if err := service.ToggleShopOpen(merchantID, req.IsOpen); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

func decorateShopTableResp(c *gin.Context, item response.ShopTableResp) response.ShopTableResp {
	if strings.HasPrefix(item.QRCodeURL, "/") {
		item.QRCodeURL = buildPublicFileURL(c, item.QRCodeURL)
	}
	return item
}

func ListMerchantShopTables(c *gin.Context) {
	merchantID := c.GetUint("userID")
	list, err := service.ListMerchantShopTables(merchantID)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	resp := make([]response.ShopTableResp, 0, len(list))
	for i := range list {
		resp = append(resp, decorateShopTableResp(c, list[i]))
	}
	utils.Success(c, resp)
}

func CreateMerchantShopTable(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.CreateShopTableReq
	if !utils.BindJSON(c, &req) {
		return
	}
	table, err := service.CreateMerchantShopTable(merchantID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, decorateShopTableResp(c, *table))
}

func UpdateMerchantShopTable(c *gin.Context) {
	merchantID := c.GetUint("userID")
	tableID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.UpdateShopTableReq
	if !utils.BindJSON(c, &req) {
		return
	}
	table, err := service.UpdateMerchantShopTable(merchantID, uint(tableID), &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, decorateShopTableResp(c, *table))
}

func DeleteMerchantShopTable(c *gin.Context) {
	merchantID := c.GetUint("userID")
	tableID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := service.DeleteMerchantShopTable(merchantID, uint(tableID)); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

func ToggleShopAutoAccept(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.ToggleShopAutoAcceptReq
	if !utils.BindJSON(c, &req) {
		return
	}
	shop, err := service.ToggleMerchantShopAutoAcceptOrders(merchantID, req.AutoAcceptOrders)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, toMerchantShopResp(shop))
}

func RotateShopPOSBindToken(c *gin.Context) {
	merchantID := c.GetUint("userID")
	shop, err := service.RotateMerchantShopPOSBindToken(merchantID)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, toMerchantShopResp(shop))
}
