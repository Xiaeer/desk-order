package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"deskorder/config"
	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/auth"
	"deskorder/pkg/geo"
	"deskorder/pkg/wechat"

	"gorm.io/gorm"
)

const (
	maxNearbyShopDistanceKm = 0.5
	maxNearbyShopCount      = 50
)

// UserLogin 用户微信登录
func UserLogin(req *request.UserLoginReq) (*response.UserLoginResp, error) {
	cfg := config.AppConfig.WeChat
	sess, err := wechat.Code2Session(cfg.UserAppID, cfg.UserSecret, req.Code)
	if err != nil {
		return nil, errors.New("微信登录失败")
	}

	user, err := repository.GetUserByOpenID(sess.OpenID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = &model.User{OpenID: sess.OpenID}
		if err := repository.CreateUser(user); err != nil {
			return nil, errors.New("创建用户失败")
		}
	} else if err != nil {
		return nil, errors.New("查询用户失败")
	}

	token, err := auth.GenerateToken(user.ID, "user")
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	return &response.UserLoginResp{Token: token}, nil
}

// GetNearbyShops 获取附近店铺（按距离排序）
func GetNearbyShops(req *request.NearbyShopsReq) ([]response.NearbyShopResp, error) {
	shops, err := repository.GetApprovedOpenShops()
	if err != nil {
		return nil, errors.New("查询店铺失败")
	}

	var list []response.NearbyShopResp
	for _, s := range shops {
		if !isValidNearbyCoordinate(s.Latitude, s.Longitude) {
			continue
		}

		dist := geo.DistanceKm(req.Latitude, req.Longitude, s.Latitude, s.Longitude)
		if dist > maxNearbyShopDistanceKm {
			continue
		}

		list = append(list, response.NearbyShopResp{
			ID:          s.ID,
			Name:        s.Name,
			Logo:        s.Logo,
			Address:     s.Address,
			Distance:    dist,
			IsOpen:      s.IsOpen,
			Description: s.Description,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Distance < list[j].Distance
	})

	if len(list) > maxNearbyShopCount {
		list = list[:maxNearbyShopCount]
	}

	return list, nil
}

func isValidNearbyCoordinate(latitude, longitude float64) bool {
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return false
	}

	return latitude != 0 || longitude != 0
}

// GetShopMenu 获取店铺菜单（分类+在售商品）
func GetShopMenu(shopID uint) (*response.MenuResp, error) {
	shop, err := repository.GetShopByID(shopID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}

	categories, err := repository.GetCategoriesByShopID(shop.ID)
	if err != nil {
		return nil, errors.New("查询分类失败")
	}

	var menu []response.CategoryWithProducts
	for _, cat := range categories {
		products, err := repository.GetOnSaleProductsByCategoryID(cat.ID)
		if err != nil {
			continue
		}
		var pList []response.ProductResp
		for _, p := range products {
			pList = append(pList, response.ProductResp{
				ID:          p.ID,
				CategoryID:  p.CategoryID,
				Name:        p.Name,
				Image:       p.Image,
				Price:       p.Price,
				Description: p.Description,
				Options:     p.Options,
				IsOnSale:    p.IsOnSale,
				Sort:        p.Sort,
			})
		}
		menu = append(menu, response.CategoryWithProducts{
			CategoryResp: response.CategoryResp{
				ID:   cat.ID,
				Name: cat.Name,
				Sort: cat.Sort,
			},
			Products: pList,
		})
	}

	return &response.MenuResp{
		ShopID:     shop.ID,
		ShopName:   shop.Name,
		Categories: menu,
	}, nil
}

// CreateOrder 用户下单
func CreateOrder(userID uint, req *request.CreateOrderReq) (*response.OrderResp, error) {
	// 校验店铺存在且营业中
	shop, err := repository.GetShopByID(req.ShopID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusApproved || !shop.IsOpen {
		return nil, errors.New("店铺未营业")
	}
	shopTable, err := resolveCreateOrderShopTable(req.ShopID, req.ShopTableID, req.EntryScene)
	if err != nil {
		return nil, err
	}

	// 构建订单项并计算总价
	var items []model.OrderItem
	var totalAmount int
	for _, item := range req.Items {
		product, err := repository.GetProductByIDForOrder(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("商品(ID:%d)不存在或已下架", item.ProductID)
		}
		if product.ShopID != req.ShopID {
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

	// 生成订单号
	orderNo := fmt.Sprintf("DO%d%d", time.Now().UnixNano(), userID)

	order := &model.Order{
		OrderNo:     orderNo,
		UserID:      userID,
		ShopID:      req.ShopID,
		TotalAmount: totalAmount,
		PayChannel:  model.PaymentChannelWechat,
		Status:      model.OrderStatusPending,
		Remark:      req.Remark,
		Items:       items,
		EntryScene:  strings.TrimSpace(req.EntryScene),
	}
	if shopTable != nil {
		shopTableID := shopTable.ID
		order.ShopTableID = &shopTableID
		order.TableNoSnapshot = shopTable.TableNo
	}

	if err := repository.CreateOrder(order); err != nil {
		return nil, errors.New("创建订单失败")
	}

	// 重新加载完整订单
	created, err := repository.GetOrderByID(order.ID)
	if err != nil {
		return nil, errors.New("查询订单失败")
	}
	refundVisible, refundEnabled, _ := GetMiniUserRefundFeatureConfig()

	return orderToResp(created, refundVisible, refundEnabled), nil
}

// GetUserOrders 获取用户订单列表
func GetUserOrders(userID uint, page, size int) ([]response.OrderResp, int64, error) {
	orders, total, err := repository.GetOrdersByUserID(userID, page, size)
	if err != nil {
		return nil, 0, errors.New("查询订单失败")
	}
	refundVisible, refundEnabled, _ := GetMiniUserRefundFeatureConfig()
	var list []response.OrderResp
	for i := range orders {
		list = append(list, *orderToResp(&orders[i], refundVisible, refundEnabled))
	}
	return list, total, nil
}

// GetUserOrderDetail 获取用户订单详情
func GetUserOrderDetail(userID, orderID uint) (*response.OrderResp, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权查看该订单")
	}
	refundVisible, refundEnabled, _ := GetMiniUserRefundFeatureConfig()
	return orderToResp(order, refundVisible, refundEnabled), nil
}

// orderToResp 订单 model 转 response
func orderToResp(o *model.Order, refundVisible, refundEnabled bool) *response.OrderResp {
	canUserCancel := o.Status == model.OrderStatusPending
	canUserRefund := refundVisible && refundEnabled && o.Status == model.OrderStatusPaid && o.PayChannel == model.PaymentChannelBalance
	refundActionText := ""
	if canUserRefund {
		refundActionText = "退款回余额"
	}
	resp := &response.OrderResp{
		ID:                 o.ID,
		OrderNo:            o.OrderNo,
		UserID:             o.UserID,
		ShopID:             o.ShopID,
		ShopTableID:        o.ShopTableID,
		TableNoSnapshot:    o.TableNoSnapshot,
		EntryScene:         o.EntryScene,
		ShopName:           o.Shop.Name,
		TotalAmount:        o.TotalAmount,
		PayChannel:         o.PayChannel,
		BalancePaidAmount:  o.BalancePaidAmount,
		WechatPaidAmount:   o.WechatPaidAmount,
		Status:             o.Status,
		Remark:             o.Remark,
		PaidAt:             o.PaidAt,
		CreatedAt:          o.CreatedAt,
		CanUserCancel:      canUserCancel,
		CanUserRefund:      canUserRefund,
		RefundEntryVisible: refundVisible,
		RefundActionText:   refundActionText,
	}
	for _, item := range o.Items {
		resp.Items = append(resp.Items, response.OrderItemResp{
			ID:              item.ID,
			ProductID:       item.ProductID,
			Name:            item.Name,
			Image:           item.Image,
			Price:           item.Price,
			SelectedOptions: item.SelectedOptions,
			OptionSummary:   item.OptionSummary,
			Quantity:        item.Quantity,
			Subtotal:        item.Subtotal,
		})
	}
	return resp
}
