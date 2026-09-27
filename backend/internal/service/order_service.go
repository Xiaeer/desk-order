package service

import (
	"errors"
	"time"

	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/internal/ws"
)

func transitionOrderToAccepted(order *model.Order) error {
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.Status == model.OrderStatusAccepted || order.Status == model.OrderStatusCompleted {
		return nil
	}
	if order.Status != model.OrderStatusPaid {
		return errors.New("订单状态不允许接单")
	}

	order.Status = model.OrderStatusAccepted
	if err := repository.UpdateOrder(order); err != nil {
		return errors.New("接单失败")
	}

	if ws.DefaultHub != nil {
		ws.DefaultHub.PushToShop(order.ShopID, ws.Message{
			Type: ws.MsgTypeOrderUpdate,
			Data: order,
		})
	}
	return nil
}

// AcceptOrder 商家接单
func AcceptOrder(merchantID uint, orderID uint) error {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	// 验证归属
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != order.ShopID {
		return errors.New("无权操作")
	}
	if order.Status != model.OrderStatusPaid {
		return errors.New("订单状态不允许接单")
	}
	return transitionOrderToAccepted(order)
}

func AutoAcceptOrder(orderID uint) error {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	if !order.Shop.AutoAcceptOrders {
		return nil
	}
	return transitionOrderToAccepted(order)
}

// CompleteOrder 完成订单
func CompleteOrder(merchantID uint, orderID uint) error {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != order.ShopID {
		return errors.New("无权操作")
	}
	if order.Status != model.OrderStatusAccepted {
		return errors.New("订单状态不允许完成")
	}

	now := time.Now()
	order.Status = model.OrderStatusCompleted
	order.PaidAt = &now
	if err := repository.UpdateOrder(order); err != nil {
		return err
	}

	if ws.DefaultHub != nil {
		ws.DefaultHub.PushToShop(order.ShopID, ws.Message{
			Type: ws.MsgTypeOrderUpdate,
			Data: order,
		})
	}
	return nil
}

// GetMerchantOrders 商家获取订单列表
func GetMerchantOrders(merchantID uint, status *int, page, size int) ([]model.Order, int64, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, 0, errors.New("店铺不存在")
	}
	return repository.GetOrdersByShopID(shop.ID, status, page, size)
}

// GetMerchantOrderDetail 商家获取订单详情
func GetMerchantOrderDetail(merchantID uint, orderID uint) (*model.Order, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != order.ShopID {
		return nil, errors.New("无权查看")
	}
	return order, nil
}
