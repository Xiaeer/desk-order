package handler

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"deskorder/internal/dto/request"
	"deskorder/internal/repository"
	"deskorder/internal/service"
	"deskorder/internal/ws"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	errPOSBindTokenMissing = errors.New("POS 密钥未初始化，请先到商户 H5 店铺页查看 POS 配置")
	errPOSBindTokenInvalid = errors.New("POS 密钥错误")
)

type posLoginReq struct {
	ShopID       uint   `json:"shop_id"`
	TerminalName string `json:"terminal_name"`
	POSToken     string `json:"pos_token"`
}

func validatePOSBinding(shopID uint, posToken string) error {
	shop, err := repository.GetShopByID(shopID)
	if err != nil {
		return errors.New("店铺不存在")
	}
	expected := strings.TrimSpace(shop.PosBindToken)
	if expected == "" {
		return errPOSBindTokenMissing
	}
	provided := strings.TrimSpace(posToken)
	if provided == "" || subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) != 1 {
		return errPOSBindTokenInvalid
	}
	return nil
}

// POSLogin POS 终端登录
func POSLogin(c *gin.Context) {
	var req posLoginReq
	if !utils.BindJSON(c, &req) {
		return
	}
	if req.ShopID == 0 {
		utils.Fail(c, 400, "shop_id 参数错误")
		return
	}
	if err := validatePOSBinding(req.ShopID, req.POSToken); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	utils.Success(c, nil)
}

// POSWebSocket POS WebSocket 连接
func POSWebSocket(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "shop_id 参数错误")
		return
	}
	if err := validatePOSBinding(uint(shopID), c.Query("pos_token")); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Hub:    ws.DefaultHub,
		ShopID: uint(shopID),
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	ws.DefaultHub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}

func POSGetMenu(c *gin.Context) {
	var req request.POSAuthReq
	if !utils.BindJSON(c, &req) {
		return
	}
	if err := validatePOSBinding(req.ShopID, req.POSToken); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	resp, err := service.GetPOSMenu(req.ShopID)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func POSCreateOrder(c *gin.Context) {
	var req request.POSCreateOrderReq
	if !utils.BindJSON(c, &req) {
		return
	}
	if err := validatePOSBinding(req.ShopID, req.POSToken); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	resp, err := service.CreatePOSOrder(req.ShopID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func POSCreateOrderPay(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.POSOrderPayReq
	if !utils.BindJSON(c, &req) {
		return
	}
	if err := validatePOSBinding(req.ShopID, req.POSToken); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	resp, err := service.CreatePOSOrderNativePay(req.ShopID, uint(orderID), c.ClientIP())
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func POSGetOrderDetail(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.POSOrderStatusReq
	if !utils.BindQuery(c, &req) {
		return
	}
	if err := validatePOSBinding(req.ShopID, req.POSToken); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	resp, err := service.GetPOSOrderDetail(req.ShopID, uint(orderID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func POSCancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.POSOrderCancelReq
	if !utils.BindJSON(c, &req) {
		return
	}
	if err := validatePOSBinding(req.ShopID, req.POSToken); err != nil {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	resp, err := service.CancelPOSPendingOrder(req.ShopID, uint(orderID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}
