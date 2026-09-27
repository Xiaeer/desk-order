package handler

import (
	"strconv"
	"strings"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/repository"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

func handleAdminShopError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}

	switch {
	case err.Error() == "店铺不存在":
		utils.Fail(c, 404, err.Error())
	case err.Error() == "该店铺已审核", err.Error() == "店铺存在订单，不能删除":
		utils.Fail(c, 400, err.Error())
	default:
		utils.Fail(c, 500, defaultMsg+": "+err.Error())
	}
}

func handleAdminMerchantError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}

	switch err.Error() {
	case "商家不存在":
		utils.Fail(c, 404, err.Error())
	case "店铺存在订单，不能删除":
		utils.Fail(c, 400, err.Error())
	default:
		utils.Fail(c, 500, defaultMsg+": "+err.Error())
	}
}

func handleAdminUserError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}
	switch err.Error() {
	case "用户不存在":
		utils.Fail(c, 404, err.Error())
	case "参数错误", "用户余额不存在", "用户当前无可清理历史余额", "用户存在多条历史余额批次，请先人工处理", "用户当前余额不足，无法清理历史余额，请人工处理":
		utils.Fail(c, 400, err.Error())
	default:
		utils.Fail(c, 500, defaultMsg+": "+err.Error())
	}
}

func handleAdminSystemConfigError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}
	if err == service.ErrInvalidSystemConfig {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Fail(c, 500, defaultMsg+": "+err.Error())
}

func handleAdminRefundError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}
	switch {
	case err.Error() == "参数错误", err.Error() == "请输入驳回原因", err.Error() == "请输入外部退款单号", err.Error() == "请输入充值订单ID或充值单号", err.Error() == "退款操作无效", err.Error() == "当前充值订单无可退本金", err.Error() == "当前充值订单已有待处理退款申请", err.Error() == "当前充值订单已有成功退款记录", err.Error() == "当前充值订单状态不可退款", err.Error() == "当前退款申请状态不可驳回", err.Error() == "当前退款申请状态不可完成", err.Error() == "当前退款申请状态不可同步", err.Error() == "当前退款执行状态不可同步", err.Error() == "用户当前余额不足，无法完成充值退款，请人工处理", err.Error() == "当前充值订单不是微信支付，请填写外部退款单号后手工完成", err.Error() == "微信退款未配置完成，请先配置退款证书或填写外部退款单号手工完成", err.Error() == "微信退款查询未配置完成，请先配置后端支付参数", err.Error() == "微信退款状态未知，请稍后重试", strings.HasPrefix(err.Error(), "微信退款失败"), strings.HasPrefix(err.Error(), "微信退款状态查询失败"):
		utils.Fail(c, 400, err.Error())
	case err.Error() == "退款申请不存在", err.Error() == "充值订单不存在", err.Error() == "充值批次不存在", err.Error() == "退款执行记录不存在":
		utils.Fail(c, 404, err.Error())
	default:
		utils.Fail(c, 500, defaultMsg+": "+err.Error())
	}
}

func handleAdminRechargeOrderError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}
	switch err.Error() {
	case "参数错误", "充值订单状态无效", "充值订单时间范围无效", "充值订单建议动作无效", "当前充值订单状态不可清理", "当前充值订单无可清理剩余余额", "当前充值订单已有待处理退款申请", "用户当前余额不足，无法作废此批次剩余，请人工处理", "该用户存在多笔未建账旧充值，无法按订单精确清理，请先人工处理历史余额", "旧充值订单未建批次，当前无可清理历史余额", "旧充值订单历史余额与订单金额不匹配，请先人工处理":
		utils.Fail(c, 400, err.Error())
	case "充值订单不存在", "充值批次不存在":
		utils.Fail(c, 404, err.Error())
	default:
		utils.Fail(c, 500, defaultMsg+": "+err.Error())
	}
}

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	var req request.AdminLoginReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminLogin(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, resp)
}

// AdminGetMerchants 获取商家列表
func AdminGetMerchants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	merchants, total, err := repository.GetMerchantList(page, size, nil)
	if err != nil {
		utils.Fail(c, 500, "获取商家列表失败")
		return
	}

	var list []response.AdminMerchantResp
	for _, m := range merchants {
		item := response.AdminMerchantResp{
			ID:        m.ID,
			Name:      m.Name,
			Phone:     m.Phone,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			HasShop:   repository.MerchantHasShop(m.ID),
		}
		shop, err := repository.GetShopByMerchantID(m.ID)
		if err == nil {
			item.ShopName = shop.Name
		}
		list = append(list, item)
	}
	utils.SuccessPage(c, list, total, page, size)
}

// AdminGetUsers 获取用户列表
func AdminGetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keyword := c.Query("keyword")
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	list, total, err := service.AdminGetUsers(page, size, keyword)
	if err != nil {
		handleAdminUserError(c, err, "获取用户列表失败")
		return
	}
	utils.SuccessPage(c, list, total, page, size)
}

// AdminGetUserDetail 获取用户详情
func AdminGetUserDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.AdminGetUserDetail(uint(id))
	if err != nil {
		handleAdminUserError(c, err, "获取用户详情失败")
		return
	}
	utils.Success(c, resp)
}

func AdminCleanupUserHistoricalBalance(c *gin.Context) {
	adminID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.AdminCleanupUserHistoricalBalanceReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminCleanupUserHistoricalBalance(adminID, uint(id), &req)
	if err != nil {
		handleAdminUserError(c, err, "清理用户历史余额失败")
		return
	}
	utils.Success(c, resp)
}

// AdminGetSystemConfigs 获取后台系统配置
func AdminGetSystemConfigs(c *gin.Context) {
	resp, err := service.GetAdminSystemConfig()
	if err != nil {
		handleAdminSystemConfigError(c, err, "获取系统配置失败")
		return
	}
	utils.Success(c, resp)
}

// AdminGetRechargeOrders 获取充值订单列表
func AdminGetRechargeOrders(c *gin.Context) {
	var req request.AdminRechargeOrderListReq
	if !utils.BindQuery(c, &req) {
		return
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}
	list, total, err := service.AdminGetRechargeOrders(page, size, req.Status, req.OrderNo, req.UserKeyword, req.CreatedStartAt, req.CreatedEndAt, req.SuggestedAction)
	if err != nil {
		handleAdminRechargeOrderError(c, err, "获取充值订单列表失败")
		return
	}
	utils.SuccessPage(c, list, total, page, size)
}

// AdminGetRechargeOrderDetail 获取充值订单详情
func AdminGetRechargeOrderDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.AdminGetRechargeOrderDetail(uint(id))
	if err != nil {
		handleAdminRechargeOrderError(c, err, "获取充值订单详情失败")
		return
	}
	utils.Success(c, resp)
}

// AdminCleanupRechargeOrderRemaining 清理充值订单剩余批次
func AdminCleanupRechargeOrderRemaining(c *gin.Context) {
	adminID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.AdminCleanupRechargeOrderReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminCleanupRechargeOrderRemaining(adminID, uint(id), &req)
	if err != nil {
		handleAdminRechargeOrderError(c, err, "作废此批次剩余失败")
		return
	}
	utils.Success(c, resp)
}

// AdminUpdateSystemConfig 更新后台系统配置
func AdminUpdateSystemConfig(c *gin.Context) {
	adminID := c.GetUint("userID")
	var req request.AdminSystemConfigReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.UpdateAdminSystemConfig(adminID, &req)
	if err != nil {
		handleAdminSystemConfigError(c, err, "更新系统配置失败")
		return
	}
	utils.Success(c, resp)
}

// AdminGetRechargeRefundRequests 获取充值退款申请列表
func AdminGetRechargeRefundRequests(c *gin.Context) {
	var req request.AdminRefundRequestListReq
	if !utils.BindQuery(c, &req) {
		return
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}
	list, total, err := service.AdminGetRechargeRefundRequests(page, size, req.Status, req.RequestChannel)
	if err != nil {
		handleAdminRefundError(c, err, "获取充值退款申请列表失败")
		return
	}
	utils.SuccessPage(c, list, total, page, size)
}

// AdminGetRechargeRefundRequestDetail 获取充值退款申请详情
func AdminGetRechargeRefundRequestDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.AdminGetRechargeRefundRequestDetail(uint(id))
	if err != nil {
		handleAdminRefundError(c, err, "获取充值退款申请详情失败")
		return
	}
	utils.Success(c, resp)
}

// AdminCreateRechargeRefundRequest 创建充值退款申请
func AdminCreateRechargeRefundRequest(c *gin.Context) {
	adminID := c.GetUint("userID")
	var req request.AdminCreateRechargeRefundReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminCreateRechargeRefundRequest(adminID, &req)
	if err != nil {
		handleAdminRefundError(c, err, "创建充值退款申请失败")
		return
	}
	utils.Success(c, resp)
}

// AdminReviewRechargeRefundRequest 驳回或完成充值退款申请
func AdminReviewRechargeRefundRequest(c *gin.Context) {
	adminID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.AdminReviewRechargeRefundReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminReviewRechargeRefundRequest(adminID, uint(id), &req)
	if err != nil {
		handleAdminRefundError(c, err, "处理充值退款申请失败")
		return
	}
	utils.Success(c, resp)
}

// AdminSyncRechargeRefundRequestStatus 同步充值退款申请微信状态
func AdminSyncRechargeRefundRequestStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	resp, err := service.AdminSyncRechargeRefundRequestStatus(uint(id))
	if err != nil {
		handleAdminRefundError(c, err, "同步充值退款状态失败")
		return
	}
	utils.Success(c, resp)
}

// AdminGetMerchantDetail 获取商家详情
func AdminGetMerchantDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	merchant, err := repository.GetMerchantByID(uint(id))
	if err != nil {
		utils.Fail(c, 404, "商家不存在")
		return
	}

	resp := response.AdminMerchantResp{
		ID:        merchant.ID,
		Name:      merchant.Name,
		Phone:     merchant.Phone,
		CreatedAt: merchant.CreatedAt.Format("2006-01-02 15:04:05"),
		HasShop:   repository.MerchantHasShop(merchant.ID),
	}
	shop, err := repository.GetShopByMerchantID(merchant.ID)
	if err == nil {
		resp.ShopName = shop.Name
	}

	utils.Success(c, resp)
}

// AdminDeleteMerchant 后台删除商家
func AdminDeleteMerchant(c *gin.Context) {
	adminID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if err := service.AdminDeleteMerchant(adminID, uint(id)); err != nil {
		handleAdminMerchantError(c, err, "删除商家失败")
		return
	}
	utils.Success(c, nil)
}

// AdminAuditShop 审核店铺
func AdminAuditShop(c *gin.Context) {
	adminID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var req request.AdminAuditShopReq
	if !utils.BindJSON(c, &req) {
		return
	}

	if err := service.AdminAuditShop(adminID, uint(id), req.Status); err != nil {
		handleAdminShopError(c, err, "审核失败")
		return
	}
	utils.Success(c, nil)
}

// AdminGetShopDetail 后台获取店铺详情
func AdminGetShopDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	resp, err := service.AdminGetShopDetail(uint(id))
	if err != nil {
		handleAdminShopError(c, err, "获取店铺详情失败")
		return
	}
	utils.Success(c, resp)
}

// AdminGetShopAuditRecords 后台获取店铺审核记录
func AdminGetShopAuditRecords(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	list, err := service.AdminGetShopAuditRecords(uint(id))
	if err != nil {
		handleAdminShopError(c, err, "获取审核记录失败")
		return
	}
	utils.Success(c, list)
}

// AdminDeleteShop 后台删除店铺
func AdminDeleteShop(c *gin.Context) {
	adminID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if err := service.AdminDeleteShop(adminID, uint(id)); err != nil {
		handleAdminShopError(c, err, "删除店铺失败")
		return
	}
	utils.Success(c, nil)
}

// AdminGetShops 后台获取店铺列表
func AdminGetShops(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}

	shops, total, err := repository.GetShopList(page, size, status)
	if err != nil {
		utils.Fail(c, 500, "获取店铺列表失败")
		return
	}

	var list []response.AdminShopResp
	for _, s := range shops {
		list = append(list, response.AdminShopResp{
			ID:           s.ID,
			Name:         s.Name,
			Logo:         s.Logo,
			Address:      s.Address,
			Phone:        s.Phone,
			Latitude:     s.Latitude,
			Longitude:    s.Longitude,
			Status:       s.Status,
			IsOpen:       s.IsOpen,
			Description:  s.Description,
			MerchantID:   s.MerchantID,
			MerchantName: s.Merchant.Name,
			CreatedAt:    s.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	utils.SuccessPage(c, list, total, page, size)
}

// AdminGetOrders 后台获取订单列表
func AdminGetOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}

	orders, total, err := repository.GetAllOrders(page, size, status)
	if err != nil {
		utils.Fail(c, 500, "获取订单列表失败")
		return
	}

	var list []response.OrderResp
	for _, o := range orders {
		item := response.OrderResp{
			ID:          o.ID,
			OrderNo:     o.OrderNo,
			UserID:      o.UserID,
			ShopID:      o.ShopID,
			ShopName:    o.Shop.Name,
			TotalAmount: o.TotalAmount,
			Status:      o.Status,
			Remark:      o.Remark,
			PaidAt:      o.PaidAt,
			CreatedAt:   o.CreatedAt,
		}
		for _, it := range o.Items {
			item.Items = append(item.Items, response.OrderItemResp{
				ID:        it.ID,
				ProductID: it.ProductID,
				Name:      it.Name,
				Image:     it.Image,
				Price:     it.Price,
				Quantity:  it.Quantity,
				Subtotal:  it.Subtotal,
			})
		}
		list = append(list, item)
	}
	utils.SuccessPage(c, list, total, page, size)
}

// AdminGetOrderDetail 后台获取订单详情
func AdminGetOrderDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	o, err := repository.GetOrderByID(uint(id))
	if err != nil {
		utils.Fail(c, 404, "订单不存在")
		return
	}

	resp := response.OrderResp{
		ID:          o.ID,
		OrderNo:     o.OrderNo,
		UserID:      o.UserID,
		ShopID:      o.ShopID,
		ShopName:    o.Shop.Name,
		TotalAmount: o.TotalAmount,
		Status:      o.Status,
		Remark:      o.Remark,
		PaidAt:      o.PaidAt,
		CreatedAt:   o.CreatedAt,
	}
	for _, it := range o.Items {
		resp.Items = append(resp.Items, response.OrderItemResp{
			ID:              it.ID,
			ProductID:       it.ProductID,
			Name:            it.Name,
			Image:           it.Image,
			Price:           it.Price,
			SelectedOptions: it.SelectedOptions,
			OptionSummary:   it.OptionSummary,
			Quantity:        it.Quantity,
			Subtotal:        it.Subtotal,
		})
	}
	utils.Success(c, resp)
}

// AdminDashboard 数据概览
func AdminDashboard(c *gin.Context) {
	resp, err := service.AdminGetDashboard()
	if err != nil {
		utils.Fail(c, 500, "获取数据概览失败")
		return
	}
	utils.Success(c, resp)
}
