package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type UnlimitedQRCodeReq struct {
	AppID      string
	Secret     string
	Scene      string
	Page       string
	EnvVersion string
	CheckPath  bool
	Width      int
}

type wechatQRCodeErrorResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func GetMiniAppUnlimitedQRCode(req *UnlimitedQRCodeReq) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("qrcode request is nil")
	}
	if strings.TrimSpace(req.Scene) == "" {
		return nil, fmt.Errorf("scene is required")
	}
	accessToken, err := GetMiniAppAccessToken(req.AppID, req.Secret)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"scene":      req.Scene,
		"check_path": req.CheckPath,
	}
	if trimmedPage := strings.TrimSpace(req.Page); trimmedPage != "" {
		payload["page"] = trimmedPage
	}
	if envVersion := strings.TrimSpace(req.EnvVersion); envVersion != "" {
		payload["env_version"] = envVersion
	}
	if req.Width > 0 {
		payload["width"] = req.Width
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal qrcode payload error: %w", err)
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", accessToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:gosec // 微信官方API地址
	if err != nil {
		return nil, fmt.Errorf("request mini app qrcode error: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read mini app qrcode response error: %w", err)
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") || len(respBody) > 0 && respBody[0] == '{' {
		var result wechatQRCodeErrorResp
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal mini app qrcode error response: %w", err)
		}
		return nil, fmt.Errorf("wechat qrcode error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return respBody, nil
}
