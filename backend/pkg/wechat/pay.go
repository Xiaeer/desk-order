package wechat

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pkcs12"
)

const unifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
const refundURL = "https://api.mch.weixin.qq.com/secapi/pay/refund"
const refundQueryURL = "https://api.mch.weixin.qq.com/pay/refundquery"

// WxPayConfig 微信支付配置（后续接入时完善）
type WxPayConfig struct {
	MchID             string
	MchAPIKey         string
	NotifyURL         string
	RefundCertP12Path string
}

// UnifiedOrderReq 统一下单请求参数
type UnifiedOrderReq struct {
	AppID      string
	MchID      string
	Body       string
	OutTradeNo string
	TotalFee   int // 单位：分
	IP         string
	NotifyURL  string
	TradeType  string // JSAPI
	OpenID     string
	ProductID  string
}

// UnifiedOrderResp 统一下单响应
type UnifiedOrderResp struct {
	PrepayID string `json:"prepay_id"`
	NonceStr string `json:"nonce_str"`
	CodeURL  string `json:"code_url"`
}

// RefundReq 微信退款请求参数
type RefundReq struct {
	AppID       string
	MchID       string
	OutTradeNo  string
	OutRefundNo string
	TotalFee    int
	RefundFee   int
	OpUserID    string
}

// RefundResp 微信退款响应
type RefundResp struct {
	RefundID      string `json:"refund_id"`
	OutRefundNo   string `json:"out_refund_no"`
	TransactionID string `json:"transaction_id"`
	ResultCode    string `json:"result_code"`
	ReturnMsg     string `json:"return_msg"`
	ErrCode       string `json:"err_code"`
	ErrCodeDes    string `json:"err_code_des"`
}

// RefundQueryReq 微信退款查询参数
type RefundQueryReq struct {
	AppID       string
	MchID       string
	OutRefundNo string
}

// RefundQueryResp 微信退款查询响应
type RefundQueryResp struct {
	RefundStatus      string `json:"refund_status"`
	RefundID          string `json:"refund_id"`
	OutRefundNo       string `json:"out_refund_no"`
	TransactionID     string `json:"transaction_id"`
	ResultCode        string `json:"result_code"`
	ReturnMsg         string `json:"return_msg"`
	ErrCode           string `json:"err_code"`
	ErrCodeDes        string `json:"err_code_des"`
	RefundSuccessTime string `json:"refund_success_time"`
}

// MiniProgramPayParams 小程序拉起支付参数
type MiniProgramPayParams struct {
	TimeStamp string
	NonceStr  string
	Package   string
	SignType  string
	PaySign   string
}

// PayNotifyResult 微信支付回调结果
type PayNotifyResult struct {
	ReturnCode string
	ResultCode string
	OutTradeNo string
}

type unifiedOrderXMLReq struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int      `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	OpenID         string   `xml:"openid,omitempty"`
	ProductID      string   `xml:"product_id,omitempty"`
}

type unifiedOrderXMLResp struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	ResultCode string   `xml:"result_code"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
	PrepayID   string   `xml:"prepay_id"`
	NonceStr   string   `xml:"nonce_str"`
	CodeURL    string   `xml:"code_url"`
}

type refundXMLReq struct {
	XMLName     xml.Name `xml:"xml"`
	AppID       string   `xml:"appid"`
	MchID       string   `xml:"mch_id"`
	NonceStr    string   `xml:"nonce_str"`
	Sign        string   `xml:"sign"`
	OutTradeNo  string   `xml:"out_trade_no"`
	OutRefundNo string   `xml:"out_refund_no"`
	TotalFee    int      `xml:"total_fee"`
	RefundFee   int      `xml:"refund_fee"`
	OpUserID    string   `xml:"op_user_id"`
}

type refundXMLResp struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code"`
	ErrCodeDes    string   `xml:"err_code_des"`
	RefundID      string   `xml:"refund_id"`
	OutRefundNo   string   `xml:"out_refund_no"`
	TransactionID string   `xml:"transaction_id"`
}

type payNotifyResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
}

type xmlMap struct {
	XMLName xml.Name   `xml:"xml"`
	Fields  []xmlField `xml:",any"`
}

type xmlField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// UnifiedOrder 统一下单
func UnifiedOrder(req *UnifiedOrderReq, apiKey string) (*UnifiedOrderResp, error) {
	nonceStr, err := generateNonceStr()
	if err != nil {
		return nil, errors.New("生成微信支付随机串失败")
	}

	params := map[string]string{
		"appid":            req.AppID,
		"mch_id":           req.MchID,
		"nonce_str":        nonceStr,
		"body":             req.Body,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        strconv.Itoa(req.TotalFee),
		"spbill_create_ip": req.IP,
		"notify_url":       req.NotifyURL,
		"trade_type":       req.TradeType,
	}
	if strings.TrimSpace(req.OpenID) != "" {
		params["openid"] = req.OpenID
	}
	if strings.TrimSpace(req.ProductID) != "" {
		params["product_id"] = req.ProductID
	}
	sign := signValues(params, apiKey)

	requestXML, err := xml.Marshal(unifiedOrderXMLReq{
		AppID:          req.AppID,
		MchID:          req.MchID,
		NonceStr:       nonceStr,
		Sign:           sign,
		Body:           req.Body,
		OutTradeNo:     req.OutTradeNo,
		TotalFee:       req.TotalFee,
		SpbillCreateIP: req.IP,
		NotifyURL:      req.NotifyURL,
		TradeType:      req.TradeType,
		OpenID:         req.OpenID,
		ProductID:      req.ProductID,
	})
	if err != nil {
		return nil, errors.New("构造微信支付请求失败")
	}

	httpResp, err := (&http.Client{Timeout: 10 * time.Second}).Post(
		unifiedOrderURL,
		"application/xml; charset=utf-8",
		bytes.NewReader(requestXML),
	)
	if err != nil {
		return nil, errors.New("请求微信支付下单失败")
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.New("读取微信支付响应失败")
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信支付下单失败: HTTP %d", httpResp.StatusCode)
	}

	var unifiedResp unifiedOrderXMLResp
	if err := xml.Unmarshal(responseBody, &unifiedResp); err != nil {
		return nil, errors.New("解析微信支付响应失败")
	}

	if unifiedResp.ReturnCode != "SUCCESS" {
		return nil, errors.New(firstNonEmpty(unifiedResp.ReturnMsg, "微信支付通信失败"))
	}
	if unifiedResp.ResultCode != "SUCCESS" {
		return nil, errors.New(firstNonEmpty(unifiedResp.ErrCodeDes, unifiedResp.ReturnMsg, "微信支付下单失败"))
	}
	tradeType := strings.ToUpper(strings.TrimSpace(req.TradeType))
	if tradeType == "NATIVE" {
		if unifiedResp.CodeURL == "" {
			return nil, errors.New("微信扫码支付预下单失败")
		}
	} else {
		if unifiedResp.PrepayID == "" {
			return nil, errors.New("微信支付预下单失败")
		}
	}

	return &UnifiedOrderResp{
		PrepayID: unifiedResp.PrepayID,
		NonceStr: unifiedResp.NonceStr,
		CodeURL:  unifiedResp.CodeURL,
	}, nil
}

// Refund 发起微信退款
func Refund(req *RefundReq, apiKey, certP12Path string) (*RefundResp, error) {
	if req == nil {
		return nil, errors.New("微信退款参数错误")
	}
	if strings.TrimSpace(certP12Path) == "" {
		return nil, errors.New("微信退款证书未配置")
	}
	if req.TotalFee <= 0 || req.RefundFee <= 0 {
		return nil, errors.New("微信退款金额错误")
	}
	if req.RefundFee > req.TotalFee {
		return nil, errors.New("微信退款金额不能大于原支付金额")
	}

	nonceStr, err := generateNonceStr()
	if err != nil {
		return nil, errors.New("生成微信退款随机串失败")
	}
	opUserID := strings.TrimSpace(req.OpUserID)
	if opUserID == "" {
		opUserID = req.MchID
	}
	params := map[string]string{
		"appid":         req.AppID,
		"mch_id":        req.MchID,
		"nonce_str":     nonceStr,
		"out_trade_no":  req.OutTradeNo,
		"out_refund_no": req.OutRefundNo,
		"total_fee":     strconv.Itoa(req.TotalFee),
		"refund_fee":    strconv.Itoa(req.RefundFee),
		"op_user_id":    opUserID,
	}
	requestXML, err := xml.Marshal(refundXMLReq{
		AppID:       req.AppID,
		MchID:       req.MchID,
		NonceStr:    nonceStr,
		Sign:        signValues(params, apiKey),
		OutTradeNo:  req.OutTradeNo,
		OutRefundNo: req.OutRefundNo,
		TotalFee:    req.TotalFee,
		RefundFee:   req.RefundFee,
		OpUserID:    opUserID,
	})
	if err != nil {
		return nil, errors.New("构造微信退款请求失败")
	}

	httpClient, err := buildRefundHTTPClient(certP12Path, req.MchID)
	if err != nil {
		return nil, err
	}
	httpResp, err := httpClient.Post(
		refundURL,
		"application/xml; charset=utf-8",
		bytes.NewReader(requestXML),
	)
	if err != nil {
		return nil, errors.New("请求微信退款失败")
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.New("读取微信退款响应失败")
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信退款失败: HTTP %d", httpResp.StatusCode)
	}

	var refundResp refundXMLResp
	if err := xml.Unmarshal(responseBody, &refundResp); err != nil {
		return nil, errors.New("解析微信退款响应失败")
	}
	if refundResp.ReturnCode != "SUCCESS" {
		return nil, errors.New(firstNonEmpty(refundResp.ReturnMsg, "微信退款通信失败"))
	}
	if refundResp.ResultCode != "SUCCESS" {
		return nil, errors.New(firstNonEmpty(refundResp.ErrCodeDes, refundResp.ReturnMsg, "微信退款申请失败"))
	}

	return &RefundResp{
		RefundID:      refundResp.RefundID,
		OutRefundNo:   refundResp.OutRefundNo,
		TransactionID: refundResp.TransactionID,
		ResultCode:    refundResp.ResultCode,
		ReturnMsg:     refundResp.ReturnMsg,
		ErrCode:       refundResp.ErrCode,
		ErrCodeDes:    refundResp.ErrCodeDes,
	}, nil
}

// RefundQuery 查询微信退款状态
func RefundQuery(req *RefundQueryReq, apiKey string) (*RefundQueryResp, error) {
	if req == nil || strings.TrimSpace(req.OutRefundNo) == "" {
		return nil, errors.New("微信退款查询参数错误")
	}
	nonceStr, err := generateNonceStr()
	if err != nil {
		return nil, errors.New("生成微信退款查询随机串失败")
	}
	params := map[string]string{
		"appid":         req.AppID,
		"mch_id":        req.MchID,
		"nonce_str":     nonceStr,
		"out_refund_no": req.OutRefundNo,
	}
	requestXML, err := xml.Marshal(struct {
		XMLName     xml.Name `xml:"xml"`
		AppID       string   `xml:"appid"`
		MchID       string   `xml:"mch_id"`
		NonceStr    string   `xml:"nonce_str"`
		Sign        string   `xml:"sign"`
		OutRefundNo string   `xml:"out_refund_no"`
	}{
		AppID:       req.AppID,
		MchID:       req.MchID,
		NonceStr:    nonceStr,
		Sign:        signValues(params, apiKey),
		OutRefundNo: req.OutRefundNo,
	})
	if err != nil {
		return nil, errors.New("构造微信退款查询请求失败")
	}
	httpResp, err := (&http.Client{Timeout: 10 * time.Second}).Post(
		refundQueryURL,
		"application/xml; charset=utf-8",
		bytes.NewReader(requestXML),
	)
	if err != nil {
		return nil, errors.New("请求微信退款查询失败")
	}
	defer httpResp.Body.Close()
	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.New("读取微信退款查询响应失败")
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信退款查询失败: HTTP %d", httpResp.StatusCode)
	}
	values, err := parseXMLMap(responseBody)
	if err != nil {
		return nil, errors.New("解析微信退款查询响应失败")
	}
	if values["return_code"] != "SUCCESS" {
		return nil, errors.New(firstNonEmpty(values["return_msg"], "微信退款查询通信失败"))
	}
	if values["result_code"] != "SUCCESS" {
		return nil, errors.New(firstNonEmpty(values["err_code_des"], values["return_msg"], "微信退款查询失败"))
	}
	refundStatus := firstNonEmpty(values["refund_status_0"], values["refund_status"])
	if refundStatus == "" {
		return nil, errors.New("微信退款查询未返回退款状态")
	}
	return &RefundQueryResp{
		RefundStatus:      refundStatus,
		RefundID:          firstNonEmpty(values["refund_id_0"], values["refund_id"]),
		OutRefundNo:       firstNonEmpty(values["out_refund_no_0"], values["out_refund_no"]),
		TransactionID:     values["transaction_id"],
		ResultCode:        values["result_code"],
		ReturnMsg:         values["return_msg"],
		ErrCode:           values["err_code"],
		ErrCodeDes:        values["err_code_des"],
		RefundSuccessTime: firstNonEmpty(values["refund_success_time_0"], values["refund_success_time"]),
	}, nil
}

// BuildMiniProgramPayParams 构造小程序拉起支付参数
func BuildMiniProgramPayParams(appID, prepayID, apiKey string) *MiniProgramPayParams {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr, err := generateNonceStr()
	if err != nil {
		nonceStr = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	packageValue := "prepay_id=" + prepayID
	params := map[string]string{
		"appId":     appID,
		"timeStamp": timeStamp,
		"nonceStr":  nonceStr,
		"package":   packageValue,
		"signType":  "MD5",
	}

	return &MiniProgramPayParams{
		TimeStamp: timeStamp,
		NonceStr:  nonceStr,
		Package:   packageValue,
		SignType:  "MD5",
		PaySign:   signValues(params, apiKey),
	}
}

// ParsePayNotify 解析微信支付回调
func ParsePayNotify(body []byte) (*PayNotifyResult, error) {
	values, err := parseXMLMap(body)
	if err != nil {
		return nil, err
	}

	return &PayNotifyResult{
		ReturnCode: values["return_code"],
		ResultCode: values["result_code"],
		OutTradeNo: values["out_trade_no"],
	}, nil
}

// VerifyNotifySign 验证支付回调签名
func VerifyNotifySign(body []byte, apiKey string) (bool, error) {
	values, err := parseXMLMap(body)
	if err != nil {
		return false, err
	}

	sign := strings.TrimSpace(values["sign"])
	if sign == "" {
		return false, errors.New("签名缺失")
	}
	delete(values, "sign")

	return signValues(values, apiKey) == sign, nil
}

// BuildPayNotifyResponse 生成微信支付回调响应 XML
func BuildPayNotifyResponse(success bool, message string) []byte {
	returnCode := "FAIL"
	if success {
		returnCode = "SUCCESS"
	}

	payload, _ := xml.Marshal(payNotifyResponse{
		ReturnCode: returnCode,
		ReturnMsg:  firstNonEmpty(message, "OK"),
	})
	return payload
}

func buildRefundHTTPClient(certP12Path, password string) (*http.Client, error) {
	keyPair, err := loadPKCS12KeyPair(certP12Path, password)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{keyPair},
		},
	}
	return &http.Client{Timeout: 15 * time.Second, Transport: transport}, nil
}

func loadPKCS12KeyPair(certP12Path, password string) (tls.Certificate, error) {
	p12Bytes, err := os.ReadFile(certP12Path)
	if err != nil {
		return tls.Certificate{}, errors.New("读取微信退款证书失败")
	}
	pemBlocks, err := pkcs12.ToPEM(p12Bytes, password)
	if err != nil {
		return tls.Certificate{}, errors.New("解析微信退款证书失败")
	}
	var certPEM bytes.Buffer
	var keyPEM bytes.Buffer
	for _, block := range pemBlocks {
		if block == nil {
			continue
		}
		switch {
		case strings.Contains(block.Type, "PRIVATE KEY"):
			if err := pem.Encode(&keyPEM, block); err != nil {
				return tls.Certificate{}, errors.New("转换微信退款证书私钥失败")
			}
		case strings.Contains(block.Type, "CERTIFICATE"):
			if err := pem.Encode(&certPEM, block); err != nil {
				return tls.Certificate{}, errors.New("转换微信退款证书失败")
			}
		}
	}
	if certPEM.Len() == 0 || keyPEM.Len() == 0 {
		return tls.Certificate{}, errors.New("微信退款证书内容不完整")
	}
	keyPair, err := tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
	if err != nil {
		return tls.Certificate{}, errors.New("加载微信退款证书失败")
	}
	return keyPair, nil
}

func parseXMLMap(body []byte) (map[string]string, error) {
	var payload xmlMap
	if err := xml.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	values := make(map[string]string, len(payload.Fields))
	for _, field := range payload.Fields {
		values[field.XMLName.Local] = strings.TrimSpace(field.Value)
	}
	return values, nil
}

func signValues(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys)+1)
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	parts = append(parts, "key="+apiKey)

	hash := md5.Sum([]byte(strings.Join(parts, "&")))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func generateNonceStr() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
