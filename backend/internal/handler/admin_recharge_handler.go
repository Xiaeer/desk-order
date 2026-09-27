package handler

import (
	"strconv"

	"deskorder/internal/dto/request"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

func handleAdminRechargeActivityError(c *gin.Context, err error, defaultMsg string) {
	if err == nil {
		return
	}

	switch err.Error() {
	case "充值活动不存在":
		utils.Fail(c, 404, err.Error())
	case "充值活动状态不能为空", "充值活动状态无效", "活动开始时间不能晚于结束时间", "充值活动已有关联充值订单，不能删除", "仅允许删除已下线活动":
		utils.Fail(c, 400, err.Error())
	default:
		utils.Fail(c, 500, defaultMsg+": "+err.Error())
	}
}

func AdminGetRechargeActivities(c *gin.Context) {
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
		value, err := strconv.Atoi(s)
		if err == nil {
			status = &value
		}
	}

	list, total, err := service.AdminGetRechargeActivities(page, size, status)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, list, total, page, size)
}

func AdminCreateRechargeActivity(c *gin.Context) {
	var req request.AdminRechargeActivityReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminCreateRechargeActivity(&req)
	if err != nil {
		handleAdminRechargeActivityError(c, err, "创建充值活动失败")
		return
	}
	utils.Success(c, resp)
}

func AdminUpdateRechargeActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	var req request.AdminRechargeActivityReq
	if !utils.BindJSON(c, &req) {
		return
	}
	resp, err := service.AdminUpdateRechargeActivity(uint(id), &req)
	if err != nil {
		handleAdminRechargeActivityError(c, err, "更新充值活动失败")
		return
	}
	utils.Success(c, resp)
}

func AdminDeleteRechargeActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := service.AdminDeleteRechargeActivity(uint(id)); err != nil {
		handleAdminRechargeActivityError(c, err, "删除充值活动失败")
		return
	}
	utils.Success(c, nil)
}
