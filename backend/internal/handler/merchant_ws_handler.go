package handler

import (
	"net/http"
	"strings"

	"deskorder/internal/repository"
	"deskorder/internal/ws"
	"deskorder/pkg/auth"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

func MerchantWebSocket(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, "未登录")
		return
	}

	claims, err := auth.ParseToken(token)
	if err != nil || claims == nil || claims.Role != "merchant" {
		utils.FailWithStatus(c, http.StatusUnauthorized, 401, "Token 无效或已过期")
		return
	}

	shop, err := repository.GetShopByMerchantID(claims.UserID)
	if err != nil || shop == nil {
		utils.FailWithStatus(c, http.StatusForbidden, 403, "店铺不存在")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Hub:    ws.DefaultHub,
		ShopID: shop.ID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	ws.DefaultHub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}
