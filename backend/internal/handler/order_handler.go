package handler

import (
	"strconv"

	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GetMerchantOrders 商家获取订单列表
func GetMerchantOrders(c *gin.Context) {
	merchantID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var statusPtr *int
	if s := c.Query("status"); s != "" {
		st, _ := strconv.Atoi(s)
		statusPtr = &st
	}

	orders, total, err := service.GetMerchantOrders(merchantID, statusPtr, page, size)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	// 转换响应
	var list []response.OrderResp
	for _, o := range orders {
		resp := orderToResp(&o)
		list = append(list, resp)
	}
	utils.SuccessPage(c, list, total, page, size)
}

// GetMerchantOrderDetail 商家获取订单详情
func GetMerchantOrderDetail(c *gin.Context) {
	merchantID := c.GetUint("userID")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	order, err := service.GetMerchantOrderDetail(merchantID, uint(orderID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, orderToResp(order))
}

// AcceptOrder 商家接单
func AcceptOrder(c *gin.Context) {
	merchantID := c.GetUint("userID")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if err := service.AcceptOrder(merchantID, uint(orderID)); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// CompleteOrder 商家完成订单
func CompleteOrder(c *gin.Context) {
	merchantID := c.GetUint("userID")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if err := service.CompleteOrder(merchantID, uint(orderID)); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// orderToResp model.Order -> response.OrderResp
func orderToResp(o *model.Order) response.OrderResp {
	resp := response.OrderResp{
		ID:              o.ID,
		OrderNo:         o.OrderNo,
		UserID:          o.UserID,
		ShopID:          o.ShopID,
		ShopTableID:     o.ShopTableID,
		TableNoSnapshot: o.TableNoSnapshot,
		EntryScene:      o.EntryScene,
		ShopName:        o.Shop.Name,
		TotalAmount:     o.TotalAmount,
		Status:          o.Status,
		Remark:          o.Remark,
		PaidAt:          o.PaidAt,
		CreatedAt:       o.CreatedAt,
	}
	for _, item := range o.Items {
		resp.Items = append(resp.Items, response.OrderItemResp{
			ID:              item.ID,
			ProductID:       item.ProductID,
			Name:            item.Name,
			Image:           item.Image,
			Price:           item.Price,
			SelectedOptions: item.SelectedOptions,
			OptionSummary:   item.OptionSummary,
			Quantity:        item.Quantity,
			Subtotal:        item.Subtotal,
		})
	}
	return resp
}
