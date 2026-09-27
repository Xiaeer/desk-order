package handler

import (
	"io"
	"net/http"
	"strings"

	"deskorder/config"
	"deskorder/internal/service"
	"deskorder/pkg/wechat"

	"github.com/gin-gonic/gin"
)

// PayNotify 微信支付回调
func PayNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(false, "读取回调失败"))
		return
	}

	apiKey := strings.TrimSpace(config.AppConfig.WeChat.MchAPIKey)
	if apiKey == "" {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(false, "支付配置缺失"))
		return
	}

	notifyResult, err := wechat.ParsePayNotify(body)
	if err != nil {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(false, "回调报文错误"))
		return
	}

	ok, err := wechat.VerifyNotifySign(body, apiKey)
	if err != nil || !ok {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(false, "回调验签失败"))
		return
	}

	if notifyResult.ReturnCode != "SUCCESS" || notifyResult.ResultCode != "SUCCESS" {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(true, "OK"))
		return
	}

	if err := service.PayNotify(notifyResult.OutTradeNo); err != nil {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(false, err.Error()))
		return
	}

	c.Data(http.StatusOK, "application/xml; charset=utf-8", wechat.BuildPayNotifyResponse(true, "OK"))
}
