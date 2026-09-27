package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Code2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type UserPhoneNumberResp struct {
	ErrCode   int           `json:"errcode"`
	ErrMsg    string        `json:"errmsg"`
	PhoneInfo UserPhoneInfo `json:"phone_info"`
}

type UserPhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

// Code2Session 用微信 code 换取 openid 和 session_key
func Code2Session(appID, secret, code string) (*Code2SessionResp, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, secret, code,
	)

	resp, err := http.Get(url) //nolint:gosec // 微信官方API地址
	if err != nil {
		return nil, fmt.Errorf("request wechat api error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	var result Code2SessionResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return &result, nil
}

// GetUserPhoneNumber 使用手机号授权 code 获取微信绑定手机号
func GetUserPhoneNumber(appID, secret, code string) (*UserPhoneNumberResp, error) {
	accessToken, err := GetMiniAppAccessToken(appID, secret)
	if err != nil {
		return nil, fmt.Errorf("get mini app access token error: %w", err)
	}
	body, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return nil, fmt.Errorf("marshal phone payload error: %w", err)
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", accessToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:gosec // 微信官方API地址
	if err != nil {
		return nil, fmt.Errorf("request wechat phone api error: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read phone response error: %w", err)
	}
	var result UserPhoneNumberResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal phone response error: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat phone error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return &result, nil
}
