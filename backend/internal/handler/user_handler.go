package handler

import (
	"io"
	"strconv"
	"strings"

	"deskorder/internal/dto/request"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

// UserLogin 用户微信登录
func UserLogin(c *gin.Context) {
	var req request.UserLoginReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.UserLogin(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetNearbyShops 获取附近商家列表
func GetNearbyShops(c *gin.Context) {
	var req request.NearbyShopsReq
	if !utils.BindQuery(c, &req) {
		return
	}
	list, err := service.GetNearbyShops(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetShopMenu 获取商家菜单
func GetShopMenu(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	menu, err := service.GetShopMenu(uint(shopID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, menu)
}

func GetUserShopTables(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	list, err := service.ListUserShopTables(uint(shopID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

func ResolveShopTableScene(c *gin.Context) {
	sceneToken := strings.TrimSpace(c.Query("scene"))
	resp, err := service.ResolveUserShopTableScene(sceneToken)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

func GetMiniUserPublicConfig(c *gin.Context) {
	resp, err := service.GetMiniUserPublicConfig()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// CreateOrder 用户下单
func CreateOrder(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.CreateOrderReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.CreateOrder(userID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserProfile 获取用户资料与余额
func GetUserProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	resp, err := service.GetUserProfile(userID)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// UpdateUserProfile 更新用户资料
func UpdateUserProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.UpdateUserProfileReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.UpdateUserProfile(userID, &req)
	if err != nil {
		switch err.Error() {
		case "用户不存在":
			utils.Fail(c, 404, err.Error())
		case "请先提供昵称或头像":
			utils.Fail(c, 400, err.Error())
		default:
			utils.Fail(c, 500, err.Error())
		}
		return
	}
	utils.Success(c, resp)
}

// BindUserPhone 绑定用户手机号
func BindUserPhone(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.BindUserPhoneReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.BindUserPhone(userID, &req)
	if err != nil {
		switch err.Error() {
		case "用户不存在":
			utils.Fail(c, 404, err.Error())
		case "微信手机号获取失败", "未获取到有效手机号":
			utils.Fail(c, 400, err.Error())
		default:
			utils.Fail(c, 500, err.Error())
		}
		return
	}
	utils.Success(c, resp)
}

// GetRechargeActivities 获取可用充值活动
func GetRechargeActivities(c *gin.Context) {
	list, err := service.GetRechargeActivities()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// CreateRechargeOrder 创建充值订单
func CreateRechargeOrder(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.CreateRechargeOrderReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.CreateRechargeOrder(userID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserRechargeOrders 获取用户充值订单列表
func GetUserRechargeOrders(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.RechargeOrderListReq
	if !utils.BindQuery(c, &req) {
		return
	}
	list, total, err := service.GetUserRechargeOrders(userID, req.Page, req.Size, req.Status)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, req.Page, req.Size)
}

// CancelUserRechargeOrder 取消待支付充值订单
func CancelUserRechargeOrder(c *gin.Context) {
	userID := c.GetUint("userID")
	rechargeOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.CancelUserRechargeOrder(userID, uint(rechargeOrderID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// ApplyUserRechargeRefund 提交用户充值退款申请
func ApplyUserRechargeRefund(c *gin.Context) {
	userID := c.GetUint("userID")
	rechargeOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.ApplyRechargeRefundReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.ApplyUserRechargeRefund(userID, uint(rechargeOrderID), req.RequestNote)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// CreateUserRechargePay 发起用户充值支付
func CreateUserRechargePay(c *gin.Context) {
	userID := c.GetUint("userID")
	rechargeOrderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.CreateUserRechargePay(userID, uint(rechargeOrderID), c.ClientIP())
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// CreateUserOrderPay 发起用户订单支付
func CreateUserOrderPay(c *gin.Context) {
	userID := c.GetUint("userID")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.OrderPayReq
	if c.Request.Body != nil {
		body, readErr := io.ReadAll(c.Request.Body)
		if readErr == nil && len(body) > 0 {
			c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
			if !utils.BindJSON(c, &req) {
				return
			}
		}
	}

	resp, err := service.CreateUserOrderPay(userID, uint(orderID), req.Method, c.ClientIP())
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}

// CancelUserOrder 取消未支付订单或将余额支付订单退款回钱包
func CancelUserOrder(c *gin.Context) {
	userID := c.GetUint("userID")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := service.CancelUserOrder(userID, uint(orderID)); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	resp, err := service.GetUserOrderDetail(userID, uint(orderID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetUserOrders 获取用户订单列表
func GetUserOrders(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.UserOrderListReq
	if !utils.BindQuery(c, &req) {
		return
	}
	list, total, err := service.GetUserOrders(userID, req.Page, req.Size)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, req.Page, req.Size)
}

// GetUserOrderDetail 获取用户订单详情
func GetUserOrderDetail(c *gin.Context) {
	userID := c.GetUint("userID")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.GetUserOrderDetail(userID, uint(orderID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// GetBalanceTransactions 获取余额流水
func GetBalanceTransactions(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.BalanceTransactionListReq
	if !utils.BindQuery(c, &req) {
		return
	}
	list, total, err := service.GetBalanceTransactions(userID, req.Page, req.Size)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, req.Page, req.Size)
}

// GetUserRefundRecords 获取用户退款记录
func GetUserRefundRecords(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.UserPageReq
	if !utils.BindQuery(c, &req) {
		return
	}
	list, total, err := service.GetUserRefundRecords(userID, req.Page, req.Size)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, req.Page, req.Size)
}

// GetUserRefundNotifications 获取用户退款通知
func GetUserRefundNotifications(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.UserPageReq
	if !utils.BindQuery(c, &req) {
		return
	}
	list, total, err := service.GetUserRefundNotifications(userID, req.Page, req.Size)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, req.Page, req.Size)
}

// ReadAllUserRefundNotifications 将退款通知全部标记已读
func ReadAllUserRefundNotifications(c *gin.Context) {
	userID := c.GetUint("userID")
	if err := service.ReadAllUserRefundNotifications(userID); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GrantUserRefundSubscribePermissions 保存用户退款订阅授权
func GrantUserRefundSubscribePermissions(c *gin.Context) {
	userID := c.GetUint("userID")
	var req request.RefundSubscribeGrantReq
	if !utils.BindJSON(c, &req) {
		return
	}
	if err := service.GrantUserRefundSubscribePermissions(userID, req.TemplateIDs); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// CancelUserRefundRequest 用户撤回退款申请
func CancelUserRefundRequest(c *gin.Context) {
	userID := c.GetUint("userID")
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.CancelUserRefundRequest(userID, uint(requestID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}
