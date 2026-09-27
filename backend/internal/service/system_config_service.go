package service

import (
	"deskorder/config"
	"strconv"
	"strings"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
)

func GetAdminSystemConfig() (*response.AdminSystemConfigResp, error) {
	visible, enabled, err := GetMiniUserRefundFeatureConfig()
	if err != nil {
		return nil, err
	}
	brandName, err := GetMiniUserBrandName()
	if err != nil {
		return nil, err
	}
	subscribeConfig, err := GetRefundSubscribeSystemConfig()
	if err != nil {
		return nil, err
	}
	return &response.AdminSystemConfigResp{
		MiniUserRefundVisible:           visible,
		MiniUserRefundApplyEnabled:      enabled,
		MiniUserBrandName:               brandName,
		RefundSubscribePage:             subscribeConfig.RefundSubscribePage,
		RefundReviewSubscribeTemplateID: subscribeConfig.RefundReviewSubscribeTemplateID,
		RefundReviewSubscribeOrderNoKey: subscribeConfig.RefundReviewSubscribeOrderNoKey,
		RefundReviewSubscribeAmountKey:  subscribeConfig.RefundReviewSubscribeAmountKey,
		RefundReviewSubscribeStatusKey:  subscribeConfig.RefundReviewSubscribeStatusKey,
		RefundReviewSubscribeRemarkKey:  subscribeConfig.RefundReviewSubscribeRemarkKey,
		RefundResultSubscribeTemplateID: subscribeConfig.RefundResultSubscribeTemplateID,
		RefundResultSubscribeOrderNoKey: subscribeConfig.RefundResultSubscribeOrderNoKey,
		RefundResultSubscribeAmountKey:  subscribeConfig.RefundResultSubscribeAmountKey,
		RefundResultSubscribeStatusKey:  subscribeConfig.RefundResultSubscribeStatusKey,
		RefundResultSubscribeRemarkKey:  subscribeConfig.RefundResultSubscribeRemarkKey,
	}, nil
}

func UpdateAdminSystemConfig(adminID uint, req *request.AdminSystemConfigReq) (*response.AdminSystemConfigResp, error) {
	if req == nil || req.MiniUserRefundVisible == nil || req.MiniUserRefundApplyEnabled == nil {
		return nil, ErrInvalidSystemConfig
	}
	refundEnabled := *req.MiniUserRefundApplyEnabled

	configs := []model.SystemConfig{
		{
			ConfigKey:   model.SystemConfigKeyMiniUserRefundVisible,
			ConfigValue: strconv.FormatBool(refundEnabled),
			ValueType:   model.SystemConfigValueTypeBool,
			Description: "兼容历史字段，始终跟随用户端自助退款申请开关",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyMiniUserRefundApplyEnabled,
			ConfigValue: strconv.FormatBool(refundEnabled),
			ValueType:   model.SystemConfigValueTypeBool,
			Description: "用户端自助退款申请开关",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyMiniUserBrandName,
			ConfigValue: normalizeMiniUserBrandName(req.MiniUserBrandName),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "用户端品牌名（用于首页与导航等品牌文案）",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundSubscribePage,
			ConfigValue: strings.TrimSpace(req.RefundSubscribePage),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款订阅消息跳转页面",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundReviewSubscribeTemplateID,
			ConfigValue: strings.TrimSpace(req.RefundReviewSubscribeTemplateID),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款审核通知模板ID",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundReviewSubscribeOrderNoKey,
			ConfigValue: strings.TrimSpace(req.RefundReviewSubscribeOrderNoKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款审核通知订单号关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundReviewSubscribeAmountKey,
			ConfigValue: strings.TrimSpace(req.RefundReviewSubscribeAmountKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款审核通知金额关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundReviewSubscribeStatusKey,
			ConfigValue: strings.TrimSpace(req.RefundReviewSubscribeStatusKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款审核通知状态关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundReviewSubscribeRemarkKey,
			ConfigValue: strings.TrimSpace(req.RefundReviewSubscribeRemarkKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款审核通知备注关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundResultSubscribeTemplateID,
			ConfigValue: strings.TrimSpace(req.RefundResultSubscribeTemplateID),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款结果通知模板ID",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundResultSubscribeOrderNoKey,
			ConfigValue: strings.TrimSpace(req.RefundResultSubscribeOrderNoKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款结果通知订单号关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundResultSubscribeAmountKey,
			ConfigValue: strings.TrimSpace(req.RefundResultSubscribeAmountKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款结果通知金额关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundResultSubscribeStatusKey,
			ConfigValue: strings.TrimSpace(req.RefundResultSubscribeStatusKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款结果通知状态关键词key",
			UpdatedBy:   adminID,
		},
		{
			ConfigKey:   model.SystemConfigKeyRefundResultSubscribeRemarkKey,
			ConfigValue: strings.TrimSpace(req.RefundResultSubscribeRemarkKey),
			ValueType:   model.SystemConfigValueTypeString,
			Description: "退款结果通知备注关键词key",
			UpdatedBy:   adminID,
		},
	}

	for i := range configs {
		if err := repository.UpsertSystemConfig(&configs[i]); err != nil {
			return nil, err
		}
	}

	return GetAdminSystemConfig()
}

func GetMiniUserPublicConfig() (*response.UserPublicConfigResp, error) {
	brandName, err := GetMiniUserBrandName()
	if err != nil {
		return nil, err
	}
	return &response.UserPublicConfigResp{MiniUserBrandName: brandName}, nil
}

func GetMiniUserBrandName() (string, error) {
	values, err := repository.GetSystemConfigsByKeys([]string{model.SystemConfigKeyMiniUserBrandName})
	if err != nil {
		return "", err
	}
	return normalizeMiniUserBrandName(values[model.SystemConfigKeyMiniUserBrandName]), nil
}

func normalizeMiniUserBrandName(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return "DeskOrder"
	}
	runes := []rune(name)
	if len(runes) > 24 {
		name = string(runes[:24])
	}
	return name
}

func GetMiniUserRefundFeatureConfig() (bool, bool, error) {
	values, err := repository.GetSystemConfigsByKeys([]string{
		model.SystemConfigKeyMiniUserRefundVisible,
		model.SystemConfigKeyMiniUserRefundApplyEnabled,
	})
	if err != nil {
		return false, false, err
	}
	enabled := parseSystemConfigBool(values[model.SystemConfigKeyMiniUserRefundApplyEnabled])
	visible := enabled
	if !enabled {
		visible = false
	}
	return visible, enabled, nil
}

type RefundSubscribeSystemConfig struct {
	RefundSubscribePage             string
	RefundReviewSubscribeTemplateID string
	RefundReviewSubscribeOrderNoKey string
	RefundReviewSubscribeAmountKey  string
	RefundReviewSubscribeStatusKey  string
	RefundReviewSubscribeRemarkKey  string
	RefundResultSubscribeTemplateID string
	RefundResultSubscribeOrderNoKey string
	RefundResultSubscribeAmountKey  string
	RefundResultSubscribeStatusKey  string
	RefundResultSubscribeRemarkKey  string
}

func GetRefundSubscribeSystemConfig() (*RefundSubscribeSystemConfig, error) {
	values, err := repository.GetSystemConfigsByKeys([]string{
		model.SystemConfigKeyRefundSubscribePage,
		model.SystemConfigKeyRefundReviewSubscribeTemplateID,
		model.SystemConfigKeyRefundReviewSubscribeOrderNoKey,
		model.SystemConfigKeyRefundReviewSubscribeAmountKey,
		model.SystemConfigKeyRefundReviewSubscribeStatusKey,
		model.SystemConfigKeyRefundReviewSubscribeRemarkKey,
		model.SystemConfigKeyRefundResultSubscribeTemplateID,
		model.SystemConfigKeyRefundResultSubscribeOrderNoKey,
		model.SystemConfigKeyRefundResultSubscribeAmountKey,
		model.SystemConfigKeyRefundResultSubscribeStatusKey,
		model.SystemConfigKeyRefundResultSubscribeRemarkKey,
	})
	if err != nil {
		return nil, err
	}
	weChatCfg := config.AppConfig.WeChat
	resp := &RefundSubscribeSystemConfig{
		RefundSubscribePage:             firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundSubscribePage]), strings.TrimSpace(weChatCfg.RefundSubscribePage), "pages/refund-center/refund-center"),
		RefundReviewSubscribeTemplateID: firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundReviewSubscribeTemplateID]), strings.TrimSpace(weChatCfg.RefundReviewSubscribeTemplateID)),
		RefundReviewSubscribeOrderNoKey: firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundReviewSubscribeOrderNoKey]), strings.TrimSpace(weChatCfg.RefundReviewSubscribeOrderNoKey)),
		RefundReviewSubscribeAmountKey:  firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundReviewSubscribeAmountKey]), strings.TrimSpace(weChatCfg.RefundReviewSubscribeAmountKey)),
		RefundReviewSubscribeStatusKey:  firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundReviewSubscribeStatusKey]), strings.TrimSpace(weChatCfg.RefundReviewSubscribeStatusKey)),
		RefundReviewSubscribeRemarkKey:  firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundReviewSubscribeRemarkKey]), strings.TrimSpace(weChatCfg.RefundReviewSubscribeRemarkKey)),
		RefundResultSubscribeTemplateID: firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundResultSubscribeTemplateID]), strings.TrimSpace(weChatCfg.RefundResultSubscribeTemplateID)),
		RefundResultSubscribeOrderNoKey: firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundResultSubscribeOrderNoKey]), strings.TrimSpace(weChatCfg.RefundResultSubscribeOrderNoKey)),
		RefundResultSubscribeAmountKey:  firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundResultSubscribeAmountKey]), strings.TrimSpace(weChatCfg.RefundResultSubscribeAmountKey)),
		RefundResultSubscribeStatusKey:  firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundResultSubscribeStatusKey]), strings.TrimSpace(weChatCfg.RefundResultSubscribeStatusKey)),
		RefundResultSubscribeRemarkKey:  firstNonEmptyString(strings.TrimSpace(values[model.SystemConfigKeyRefundResultSubscribeRemarkKey]), strings.TrimSpace(weChatCfg.RefundResultSubscribeRemarkKey)),
	}
	return resp, nil
}

func parseSystemConfigBool(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "true" || value == "1" || value == "yes" || value == "on"
}
