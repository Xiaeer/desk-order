package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"deskorder/config"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/wechat"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type refundSubscribeTemplateConfig struct {
	Scene      string
	TemplateID string
	Page       string
	OrderNoKey string
	AmountKey  string
	StatusKey  string
	RemarkKey  string
}

func GetRefundSubscribeTemplateIDs() []string {
	configs := listRefundSubscribeTemplateConfigs()
	result := make([]string, 0, len(configs))
	seen := make(map[string]struct{}, len(configs))
	for _, item := range configs {
		if item.TemplateID == "" {
			continue
		}
		if _, ok := seen[item.TemplateID]; ok {
			continue
		}
		seen[item.TemplateID] = struct{}{}
		result = append(result, item.TemplateID)
	}
	return result
}

func GrantUserRefundSubscribePermissions(userID uint, templateIDs []string) error {
	if userID == 0 || len(templateIDs) == 0 {
		return nil
	}
	configMap := make(map[string]refundSubscribeTemplateConfig)
	for _, item := range listRefundSubscribeTemplateConfigs() {
		if item.TemplateID == "" {
			continue
		}
		configMap[item.TemplateID] = item
	}
	items := make([]model.UserSubscribeMessageGrant, 0, len(templateIDs))
	seen := make(map[string]struct{}, len(templateIDs))
	now := time.Now()
	for _, templateID := range templateIDs {
		templateID = strings.TrimSpace(templateID)
		if templateID == "" {
			continue
		}
		if _, ok := seen[templateID]; ok {
			continue
		}
		seen[templateID] = struct{}{}
		cfg, ok := configMap[templateID]
		if !ok {
			continue
		}
		items = append(items, model.UserSubscribeMessageGrant{
			UserID:     userID,
			Scene:      cfg.Scene,
			TemplateID: cfg.TemplateID,
			Status:     model.SubscribeGrantStatusPending,
			AcceptedAt: now,
		})
	}
	return repository.CreateUserSubscribeMessageGrants(items)
}

func triggerRefundReviewSubscribeMessage(item *model.RefundRequest) {
	if item == nil {
		return
	}
	switch item.Status {
	case model.RefundStatusRejected, model.RefundStatusProcessing:
		trySendRefundSubscribeMessage(item, model.SubscribeSceneRefundReview)
	}
}

func triggerRefundResultSubscribeMessage(item *model.RefundRequest) {
	if item == nil {
		return
	}
	switch item.Status {
	case model.RefundStatusSuccess, model.RefundStatusFailed:
		trySendRefundSubscribeMessage(item, model.SubscribeSceneRefundResult)
	}
}

func trySendRefundSubscribeMessage(item *model.RefundRequest, scene string) {
	if item == nil || scene == "" {
		return
	}
	if err := sendRefundSubscribeMessage(item, scene); err != nil {
		log.Warn().Err(err).Uint("refund_request_id", item.ID).Str("scene", scene).Msg("failed to send refund subscribe message")
	}
}

func sendRefundSubscribeMessage(item *model.RefundRequest, scene string) error {
	if item == nil {
		return nil
	}
	cfg, ok := getRefundSubscribeTemplateConfig(scene)
	if !ok {
		return nil
	}
	grant, err := repository.GetOldestPendingUserSubscribeMessageGrant(item.UserID, scene)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	user, err := repository.GetUserByID(item.UserID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(user.OpenID) == "" {
		return nil
	}
	payload, err := buildRefundSubscribeData(item, cfg)
	if err != nil {
		return err
	}
	_, sendErr := wechat.SendSubscribeMessage(&wechat.SubscribeMessageReq{
		AppID:      config.AppConfig.WeChat.UserAppID,
		Secret:     config.AppConfig.WeChat.UserSecret,
		OpenID:     user.OpenID,
		TemplateID: cfg.TemplateID,
		Page:       cfg.Page,
		Data:       payload,
	})
	now := time.Now()
	grant.RelatedType = item.SourceType
	grant.RelatedID = item.SourceID
	if sendErr != nil {
		grant.Status = model.SubscribeGrantStatusFailed
		grant.ErrorMessage = truncateRefundSubscribeText(sendErr.Error(), 256)
		grant.ConsumedAt = &now
		return repository.UpdateUserSubscribeMessageGrant(grant)
	}
	grant.Status = model.SubscribeGrantStatusConsumed
	grant.ErrorMessage = ""
	grant.ConsumedAt = &now
	return repository.UpdateUserSubscribeMessageGrant(grant)
}

func listRefundSubscribeTemplateConfigs() []refundSubscribeTemplateConfig {
	if config.AppConfig == nil {
		return nil
	}
	systemConfig, err := GetRefundSubscribeSystemConfig()
	if err != nil || systemConfig == nil {
		return nil
	}
	return []refundSubscribeTemplateConfig{
		{
			Scene:      model.SubscribeSceneRefundReview,
			TemplateID: strings.TrimSpace(systemConfig.RefundReviewSubscribeTemplateID),
			Page:       systemConfig.RefundSubscribePage,
			OrderNoKey: strings.TrimSpace(systemConfig.RefundReviewSubscribeOrderNoKey),
			AmountKey:  strings.TrimSpace(systemConfig.RefundReviewSubscribeAmountKey),
			StatusKey:  strings.TrimSpace(systemConfig.RefundReviewSubscribeStatusKey),
			RemarkKey:  strings.TrimSpace(systemConfig.RefundReviewSubscribeRemarkKey),
		},
		{
			Scene:      model.SubscribeSceneRefundResult,
			TemplateID: strings.TrimSpace(systemConfig.RefundResultSubscribeTemplateID),
			Page:       systemConfig.RefundSubscribePage,
			OrderNoKey: strings.TrimSpace(systemConfig.RefundResultSubscribeOrderNoKey),
			AmountKey:  strings.TrimSpace(systemConfig.RefundResultSubscribeAmountKey),
			StatusKey:  strings.TrimSpace(systemConfig.RefundResultSubscribeStatusKey),
			RemarkKey:  strings.TrimSpace(systemConfig.RefundResultSubscribeRemarkKey),
		},
	}
}

func getRefundSubscribeTemplateConfig(scene string) (refundSubscribeTemplateConfig, bool) {
	for _, item := range listRefundSubscribeTemplateConfigs() {
		if item.Scene == scene && item.TemplateID != "" {
			return item, true
		}
	}
	return refundSubscribeTemplateConfig{}, false
}

func buildRefundSubscribeData(item *model.RefundRequest, cfg refundSubscribeTemplateConfig) (map[string]wechat.SubscribeMessageDataItem, error) {
	if item == nil {
		return nil, fmt.Errorf("refund request is nil")
	}
	orderNo := "-"
	if item.SourceType == model.RefundSourceTypeRecharge {
		order, err := repository.GetRechargeOrderByID(item.SourceID)
		if err != nil {
			return nil, err
		}
		orderNo = order.OrderNo
	}
	amount := item.ApprovedTotalAmount
	if amount <= 0 {
		amount = item.RequestedTotalAmount
	}
	statusText := refundSubscribeStatusText(item, cfg.Scene)
	remark := refundSubscribeRemarkText(item, cfg.Scene)
	data := make(map[string]wechat.SubscribeMessageDataItem)
	if cfg.OrderNoKey != "" {
		data[cfg.OrderNoKey] = wechat.SubscribeMessageDataItem{Value: truncateRefundSubscribeText(orderNo, 32)}
	}
	if cfg.AmountKey != "" {
		data[cfg.AmountKey] = wechat.SubscribeMessageDataItem{Value: fmt.Sprintf("¥%.2f", float64(amount)/100)}
	}
	if cfg.StatusKey != "" {
		data[cfg.StatusKey] = wechat.SubscribeMessageDataItem{Value: truncateRefundSubscribeText(statusText, 10)}
	}
	if cfg.RemarkKey != "" {
		data[cfg.RemarkKey] = wechat.SubscribeMessageDataItem{Value: truncateRefundSubscribeText(remark, 20)}
	}
	return data, nil
}

func refundSubscribeStatusText(item *model.RefundRequest, scene string) string {
	if item == nil {
		return "退款通知"
	}
	switch scene {
	case model.SubscribeSceneRefundReview:
		if item.Status == model.RefundStatusRejected {
			return "审核未通过"
		}
		if item.Status == model.RefundStatusProcessing {
			return "审核通过"
		}
	case model.SubscribeSceneRefundResult:
		if item.Status == model.RefundStatusSuccess {
			return "退款成功"
		}
		if item.Status == model.RefundStatusFailed {
			return "退款失败"
		}
	}
	return userRefundStatusText(item.Status)
}

func refundSubscribeRemarkText(item *model.RefundRequest, scene string) string {
	if item == nil {
		return "请进入退款中心查看详情"
	}
	if scene == model.SubscribeSceneRefundReview && item.Status == model.RefundStatusRejected {
		return firstNonEmptyString(item.RejectReason, item.ReviewNote, "本次退款申请未通过审核")
	}
	if scene == model.SubscribeSceneRefundReview && item.Status == model.RefundStatusProcessing {
		return "退款审核已通过，正在按原支付路径处理"
	}
	return buildUserRefundResultText(item)
}

func truncateRefundSubscribeText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || value == "" {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
