package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"deskorder/pkg/database"
)

const (
	wechatAccessTokenCachePrefix = "wechat:access_token:"
	defaultAccessTokenTTL        = 90 * time.Minute
)

type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type SubscribeMessageDataItem struct {
	Value string `json:"value"`
}

type SubscribeMessageReq struct {
	AppID      string
	Secret     string
	OpenID     string
	TemplateID string
	Page       string
	Data       map[string]SubscribeMessageDataItem
	Lang       string
}

type SubscribeMessageResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func GetMiniAppAccessToken(appID, secret string) (string, error) {
	if appID == "" || secret == "" {
		return "", fmt.Errorf("missing mini app credentials")
	}
	cacheKey := wechatAccessTokenCachePrefix + appID
	ctx := context.Background()
	if database.RDB != nil {
		cached, err := database.RDB.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			return cached, nil
		}
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, secret)
	resp, err := http.Get(url) //nolint:gosec // 微信官方API地址
	if err != nil {
		return "", fmt.Errorf("request wechat access token error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read access token response error: %w", err)
	}

	var result AccessTokenResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal access token response error: %w", err)
	}
	if result.ErrCode != 0 || result.AccessToken == "" {
		return "", fmt.Errorf("wechat access token error: %d %s", result.ErrCode, result.ErrMsg)
	}

	ttl := defaultAccessTokenTTL
	if result.ExpiresIn > 600 {
		ttl = time.Duration(result.ExpiresIn-300) * time.Second
	}
	if database.RDB != nil {
		_ = database.RDB.Set(ctx, cacheKey, result.AccessToken, ttl).Err()
	}
	return result.AccessToken, nil
}

func SendSubscribeMessage(req *SubscribeMessageReq) (*SubscribeMessageResp, error) {
	if req == nil {
		return nil, fmt.Errorf("subscribe message request is nil")
	}
	if req.OpenID == "" || req.TemplateID == "" {
		return nil, fmt.Errorf("missing subscribe message receiver or template")
	}
	accessToken, err := GetMiniAppAccessToken(req.AppID, req.Secret)
	if err != nil {
		return nil, err
	}
	lang := req.Lang
	if lang == "" {
		lang = "zh_CN"
	}
	payload := map[string]any{
		"touser":      req.OpenID,
		"template_id": req.TemplateID,
		"data":        req.Data,
		"lang":        lang,
	}
	if req.Page != "" {
		payload["page"] = req.Page
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal subscribe message payload error: %w", err)
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:gosec // 微信官方API地址
	if err != nil {
		return nil, fmt.Errorf("request subscribe message api error: %w", err)
	}
	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read subscribe message response error: %w", err)
	}
	var result SubscribeMessageResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal subscribe message response error: %w", err)
	}
	if result.ErrCode != 0 {
		return &result, fmt.Errorf("wechat subscribe message error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return &result, nil
}
