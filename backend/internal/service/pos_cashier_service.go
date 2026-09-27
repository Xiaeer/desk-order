package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"deskorder/config"
	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/wechat"
)

const posEntryScene = "pos_cashier"

func GetPOSMenu(shopID uint) (*response.MenuResp, error) {
	return GetShopMenu(shopID)
}

func CreatePOSOrder(shopID uint, req *request.POSCreateOrderReq) (*response.OrderResp, error) {
	if req == nil {
		return nil, errors.New("参数错误")
	}
	shop, err := repository.GetShopByID(shopID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusApproved || !shop.IsOpen {
		return nil, errors.New("店铺未营业")
	}
	clientRequestID := strings.TrimSpace(req.ClientRequestID)
	if clientRequestID == "" {
		return nil, errors.New("client_request_id 不能为空")
	}
	if existing, err := repository.GetOrderByShopIDAndClientRequestID(shopID, clientRequestID); err == nil && existing != nil {
		return orderToResp(existing, false, false), nil
	}

	var items []model.OrderItem
	totalAmount := 0
	for _, item := range req.Items {
		product, err := repository.GetProductByIDForOrder(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("商品(ID:%d)不存在或已下架", item.ProductID)
		}
		if product.ShopID != shopID {
			return nil, fmt.Errorf("商品(ID:%d)不属于该店铺", item.ProductID)
		}
		selectedOptions, optionSummary, optionDelta, err := resolveOrderItemSelectedOptions(product, item.SelectedOptions)
		if err != nil {
			return nil, err
		}
		finalUnitPrice := product.Price + optionDelta
		subtotal := finalUnitPrice * item.Quantity
		items = append(items, model.OrderItem{
			ProductID:       product.ID,
			Name:            product.Name,
			Image:           product.Image,
			Price:           finalUnitPrice,
			SelectedOptions: selectedOptions,
			OptionSummary:   optionSummary,
			Quantity:        item.Quantity,
			Subtotal:        subtotal,
		})
		totalAmount += subtotal
	}

	orderNo := fmt.Sprintf("PO%d%d", time.Now().UnixNano(), shopID)
	tableNo := strings.TrimSpace(req.TableNo)
	terminalRef := strings.TrimSpace(req.TerminalRef)
	entryScene := posEntryScene
	if terminalRef != "" {
		entryScene = posEntryScene + ":" + terminalRef
	}

	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          0,
		ShopID:          shopID,
		ClientRequestID: &clientRequestID,
		TableNoSnapshot: tableNo,
		EntryScene:      entryScene,
		TotalAmount:     totalAmount,
		PayChannel:      model.PaymentChannelWechat,
		Status:          model.OrderStatusPending,
		Remark:          strings.TrimSpace(req.Remark),
		Items:           items,
	}

	if err := repository.CreateOrder(order); err != nil {
		if existing, lookupErr := repository.GetOrderByShopIDAndClientRequestID(shopID, clientRequestID); lookupErr == nil && existing != nil {
			return orderToResp(existing, false, false), nil
		}
		return nil, errors.New("创建订单失败")
	}

	created, err := repository.GetOrderByID(order.ID)
	if err != nil {
		return nil, errors.New("查询订单失败")
	}
	return orderToResp(created, false, false), nil
}

func GetPOSOrderDetail(shopID, orderID uint) (*response.OrderResp, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.ShopID != shopID {
		return nil, errors.New("无权查看该订单")
	}
	return orderToResp(order, false, false), nil
}

func CancelPOSPendingOrder(shopID, orderID uint) (*response.OrderResp, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.ShopID != shopID {
		return nil, errors.New("无权取消该订单")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("当前订单状态不可取消")
	}
	order.Status = model.OrderStatusCancelled
	if err := repository.UpdateOrder(order); err != nil {
		return nil, errors.New("取消订单失败")
	}
	updated, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("查询订单失败")
	}
	return orderToResp(updated, false, false), nil
}

func CreatePOSOrderNativePay(shopID, orderID uint, clientIP string) (*response.POSNativePayResp, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.ShopID != shopID {
		return nil, errors.New("无权支付该订单")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("当前订单状态不可支付")
	}

	cfg := config.AppConfig.WeChat
	if normalizePayMode(cfg.PayMode) == payModeMock {
		mockResult := normalizeMockPayResult(cfg.MockPayResult)
		if mockResult == mockPayResultSuccess {
			if err := PayNotify(order.OrderNo); err != nil {
				return nil, err
			}
			order, _ = repository.GetOrderByID(orderID)
		}
		return &response.POSNativePayResp{
			OrderID:     order.ID,
			OrderNo:     order.OrderNo,
			TotalAmount: order.TotalAmount,
			Status:      order.Status,
			Mode:        payModeMock,
			MockResult:  mockResult,
		}, nil
	}

	payAppID := strings.TrimSpace(cfg.MerchantAppID)
	if payAppID == "" {
		payAppID = strings.TrimSpace(cfg.UserAppID)
	}
	if payAppID == "" || cfg.MchID == "" || cfg.MchAPIKey == "" || cfg.NotifyURL == "" {
		return nil, errors.New("微信支付未配置完成，请先配置商户号、API密钥、应用AppID和回调地址")
	}

	payResp, err := wechat.UnifiedOrder(&wechat.UnifiedOrderReq{
		AppID:      payAppID,
		MchID:      cfg.MchID,
		Body:       buildPayBody(order.Shop.Name),
		OutTradeNo: order.OrderNo,
		TotalFee:   order.TotalAmount,
		IP:         normalizePayClientIP(clientIP),
		NotifyURL:  cfg.NotifyURL,
		TradeType:  "NATIVE",
		ProductID:  order.OrderNo,
	}, cfg.MchAPIKey)
	if err != nil {
		return nil, err
	}

	return &response.POSNativePayResp{
		OrderID:     order.ID,
		OrderNo:     order.OrderNo,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		CodeURL:     payResp.CodeURL,
		Mode:        payModeReal,
	}, nil
}
