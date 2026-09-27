package service

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"deskorder/config"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/internal/ws"
	"deskorder/pkg/database"
	"deskorder/pkg/wechat"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	payModeMock    = "mock"
	payModeReal    = "real"
	payModeBalance = "balance"

	mockPayResultSuccess = "success"
	mockPayResultCancel  = "cancel"
	mockPayResultFail    = "fail"
)

// CreateUserOrderPay 发起用户订单支付
func CreateUserOrderPay(userID, orderID uint, method, clientIP string) (*response.MiniProgramPayResp, error) {
	payMethod := normalizeOrderPayMethod(method)
	if payMethod == model.PaymentChannelBalance {
		return payOrderByBalance(userID, orderID)
	}
	return createUserOrderWechatPay(userID, orderID, clientIP)
}

// CancelUserOrder 取消未支付订单或将余额支付订单退款回钱包
func CancelUserOrder(userID, orderID uint) error {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.UserID != userID {
		return errors.New("无权操作该订单")
	}
	if order.Status == model.OrderStatusPaid && order.PayChannel == model.PaymentChannelBalance {
		_, refundEnabled, cfgErr := GetMiniUserRefundFeatureConfig()
		if cfgErr != nil {
			return errors.New("查询系统配置失败")
		}
		if !refundEnabled {
			return errors.New("当前暂不支持自助退款")
		}
	}

	shouldPushUpdate := false
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		currentOrder, err := repository.GetOrderByIDWithDB(tx, orderID)
		if err != nil {
			return err
		}
		if currentOrder.UserID != userID {
			return errors.New("无权操作该订单")
		}

		switch {
		case currentOrder.Status == model.OrderStatusPending:
			currentOrder.Status = model.OrderStatusCancelled
			return repository.UpdateOrderWithDB(tx, currentOrder)
		case currentOrder.Status == model.OrderStatusPaid && currentOrder.PayChannel == model.PaymentChannelBalance:
			balance, err := repository.GetUserBalanceByUserIDWithDB(tx, userID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				balance = &model.UserBalance{UserID: userID}
				if err := tx.Create(balance).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			allocations, err := repository.GetBalanceConsumptionAllocationsByOrderIDWithDB(tx, orderID)
			if err != nil {
				return err
			}
			principalRefundAmount := currentOrder.TotalAmount
			giftRefundAmount := 0
			if len(allocations) > 0 {
				principalRefundAmount = 0
				giftRefundAmount = 0
				batchIDs := uniqueRechargeBatchIDs(allocations)
				batches, batchErr := repository.GetRechargeBatchesByIDsWithDB(tx, batchIDs)
				if batchErr != nil {
					return batchErr
				}
				batchMap := make(map[uint]*model.RechargeBatch, len(batches))
				for i := range batches {
					batchMap[batches[i].ID] = &batches[i]
				}
				for _, allocation := range allocations {
					principalRefundAmount += allocation.PrincipalAmount
					giftRefundAmount += allocation.GiftAmount
					batch := batchMap[allocation.RechargeBatchID]
					if batch == nil {
						return errors.New("余额分摊批次不存在")
					}
					batch.PrincipalRemaining += allocation.PrincipalAmount
					batch.GiftRemaining += allocation.GiftAmount
				}
				for _, batch := range batchMap {
					if err := repository.SaveRechargeBatchWithDB(tx, batch); err != nil {
						return err
					}
				}
				if err := repository.DeleteBalanceConsumptionAllocationsByOrderIDWithDB(tx, orderID); err != nil {
					return err
				}
			}

			balance.BalanceAmount += currentOrder.TotalAmount
			if balance.TotalConsumeAmount >= currentOrder.TotalAmount {
				balance.TotalConsumeAmount -= currentOrder.TotalAmount
			} else {
				balance.TotalConsumeAmount = 0
			}
			if err := repository.SaveUserBalanceWithDB(tx, balance); err != nil {
				return err
			}

			currentOrder.Status = model.OrderStatusCancelled
			if err := repository.UpdateOrderWithDB(tx, currentOrder); err != nil {
				return err
			}
			shopID := currentOrder.ShopID
			now := time.Now()

			refundRequest := &model.RefundRequest{
				RequestNo:                fmt.Sprintf("RR%d%d", time.Now().UnixNano(), currentOrder.ID),
				SourceType:               model.RefundSourceTypeOrder,
				SourceID:                 currentOrder.ID,
				UserID:                   userID,
				ShopID:                   &shopID,
				RequestChannel:           model.RefundRequestChannelUser,
				ReviewScope:              model.RefundReviewScopeOrder,
				Status:                   model.RefundStatusSuccess,
				RequestedTotalAmount:     currentOrder.TotalAmount,
				RequestedPrincipalAmount: principalRefundAmount,
				RequestedGiftVoidAmount:  giftRefundAmount,
				ApprovedTotalAmount:      currentOrder.TotalAmount,
				ApprovedPrincipalAmount:  principalRefundAmount,
				ApprovedGiftVoidAmount:   giftRefundAmount,
				ReviewNote:               "余额支付订单自助退款回钱包",
				ReviewerRole:             "system",
				ReviewerID:               0,
				ReviewedAt:               &now,
			}
			if err := repository.CreateRefundRequestWithDB(tx, refundRequest); err != nil {
				return err
			}

			refundExecution := &model.RefundExecution{
				RefundRequestID:        refundRequest.ID,
				ExecutionNo:            fmt.Sprintf("RE%d%d", time.Now().UnixNano(), currentOrder.ID),
				Channel:                model.RefundExecutionChannelWallet,
				OriginalPaySourceType:  model.RefundSourceTypeOrder,
				OriginalPaySourceID:    currentOrder.ID,
				ExecuteTotalAmount:     currentOrder.TotalAmount,
				ExecutePrincipalAmount: principalRefundAmount,
				ExecuteGiftVoidAmount:  giftRefundAmount,
				Status:                 model.RefundExecutionStatusSuccess,
				StartedAt:              &now,
				FinishedAt:             &now,
			}
			if err := repository.CreateRefundExecutionWithDB(tx, refundExecution); err != nil {
				return err
			}

			if err := repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
				UserID:                userID,
				ChangeAmount:          currentOrder.TotalAmount,
				PrincipalChangeAmount: principalRefundAmount,
				GiftChangeAmount:      giftRefundAmount,
				BalanceAfter:          balance.BalanceAmount,
				BizType:               model.BalanceTransactionTypeRefund,
				SourceType:            model.BalanceTransactionSourceOrder,
				SourceID:              currentOrder.ID,
				ShopID:                &shopID,
				Remark:                "余额支付订单退款回钱包",
			}); err != nil {
				return err
			}
			if err := createUserRefundNotificationWithDB(tx, userID, model.RefundSourceTypeOrder, currentOrder.ID, "订单退款已到账", "余额支付订单已退回钱包余额，可在余额流水中查看。"); err != nil {
				return err
			}
			shouldPushUpdate = true
			return nil
		case currentOrder.Status == model.OrderStatusPaid:
			return errors.New("当前订单仅支持余额支付退款")
		default:
			return errors.New("当前订单状态不可取消或退款")
		}
	}); err != nil {
		return err
	}

	if shouldPushUpdate {
		updatedOrder, err := repository.GetOrderByID(orderID)
		if err == nil && ws.DefaultHub != nil {
			ws.DefaultHub.PushToShop(updatedOrder.ShopID, ws.Message{
				Type: ws.MsgTypeOrderUpdate,
				Data: updatedOrder,
			})
		}
	}
	return nil
}

// CreateUserRechargePay 发起充值支付
func CreateUserRechargePay(userID, rechargeOrderID uint, clientIP string) (*response.MiniProgramPayResp, error) {
	rechargeOrder, err := repository.GetRechargeOrderByUserIDAndID(userID, rechargeOrderID)
	if err != nil {
		return nil, errors.New("充值订单不存在")
	}
	if rechargeOrder.Status != model.RechargeOrderStatusPending {
		return nil, errors.New("当前充值订单状态不可支付")
	}
	if err := validateRechargeOrderBeforePay(rechargeOrder, time.Now()); err != nil {
		return nil, err
	}

	cfg := config.AppConfig.WeChat
	if normalizePayMode(cfg.PayMode) == payModeMock {
		return createMockRechargePay(rechargeOrder, normalizeMockPayResult(cfg.MockPayResult))
	}
	if cfg.UserAppID == "" || cfg.MchID == "" || cfg.MchAPIKey == "" || cfg.NotifyURL == "" {
		return nil, errors.New("微信支付未配置完成，请先配置商户号、API密钥和回调地址")
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if strings.TrimSpace(user.OpenID) == "" {
		return nil, errors.New("用户微信身份缺失，请重新登录")
	}

	payResp, err := wechat.UnifiedOrder(&wechat.UnifiedOrderReq{
		AppID:      cfg.UserAppID,
		MchID:      cfg.MchID,
		Body:       buildRechargePayBody(rechargeOrder.ActivityNameSnapshot),
		OutTradeNo: rechargeOrder.OrderNo,
		TotalFee:   rechargeOrder.PayAmount,
		IP:         normalizePayClientIP(clientIP),
		NotifyURL:  cfg.NotifyURL,
		TradeType:  "JSAPI",
		OpenID:     user.OpenID,
	}, cfg.MchAPIKey)
	if err != nil {
		return nil, err
	}

	miniPayParams := wechat.BuildMiniProgramPayParams(cfg.UserAppID, payResp.PrepayID, cfg.MchAPIKey)
	return &response.MiniProgramPayResp{
		TimeStamp: miniPayParams.TimeStamp,
		NonceStr:  miniPayParams.NonceStr,
		Package:   miniPayParams.Package,
		SignType:  miniPayParams.SignType,
		PaySign:   miniPayParams.PaySign,
	}, nil
}

func validateRechargeOrderBeforePay(order *model.RechargeOrder, now time.Time) error {
	if order == nil {
		return errors.New("充值订单不存在")
	}
	if order.ExpiresAt != nil && now.After(*order.ExpiresAt) {
		return errors.New("当前充值订单已过期，请重新创建")
	}
	activity, err := repository.GetRechargeActivityByID(order.ActivityID)
	if err != nil {
		return errors.New("充值活动已更新，请重新创建订单")
	}
	if normalizeRechargeActivityRuleVersion(activity.RuleVersion) != normalizeRechargeActivityRuleVersion(order.ActivityVersionSnapshot) {
		return errors.New("充值活动已更新，请重新创建订单")
	}
	if !isRechargeActivityAvailable(activity, now) {
		return errors.New("充值活动已更新，请重新创建订单")
	}
	return nil
}

func createUserOrderWechatPay(userID, orderID uint, clientIP string) (*response.MiniProgramPayResp, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权支付该订单")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("当前订单状态不可支付")
	}

	cfg := config.AppConfig.WeChat
	if normalizePayMode(cfg.PayMode) == payModeMock {
		return createMockUserOrderPay(order, normalizeMockPayResult(cfg.MockPayResult))
	}
	if cfg.UserAppID == "" || cfg.MchID == "" || cfg.MchAPIKey == "" || cfg.NotifyURL == "" {
		return nil, errors.New("微信支付未配置完成，请先配置商户号、API密钥和回调地址")
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if strings.TrimSpace(user.OpenID) == "" {
		return nil, errors.New("用户微信身份缺失，请重新登录")
	}

	payResp, err := wechat.UnifiedOrder(&wechat.UnifiedOrderReq{
		AppID:      cfg.UserAppID,
		MchID:      cfg.MchID,
		Body:       buildPayBody(order.Shop.Name),
		OutTradeNo: order.OrderNo,
		TotalFee:   order.TotalAmount,
		IP:         normalizePayClientIP(clientIP),
		NotifyURL:  cfg.NotifyURL,
		TradeType:  "JSAPI",
		OpenID:     user.OpenID,
	}, cfg.MchAPIKey)
	if err != nil {
		return nil, err
	}

	miniPayParams := wechat.BuildMiniProgramPayParams(cfg.UserAppID, payResp.PrepayID, cfg.MchAPIKey)
	return &response.MiniProgramPayResp{
		TimeStamp: miniPayParams.TimeStamp,
		NonceStr:  miniPayParams.NonceStr,
		Package:   miniPayParams.Package,
		SignType:  miniPayParams.SignType,
		PaySign:   miniPayParams.PaySign,
		Mode:      model.PaymentChannelWechat,
	}, nil
}

// PayNotify 支付回调处理：更新订单状态或充值入账
func PayNotify(orderNo string) error {
	if strings.HasPrefix(orderNo, "RO") {
		return completeRechargePay(orderNo)
	}
	return completeWechatOrderPay(orderNo)
}

func completeWechatOrderPay(orderNo string) error {
	order, err := repository.GetOrderByOrderNo(orderNo)
	if err != nil {
		return errors.New("订单不存在")
	}

	if order.Status == model.OrderStatusPaid || order.Status == model.OrderStatusAccepted || order.Status == model.OrderStatusCompleted {
		return nil
	}
	if order.Status != model.OrderStatusPending {
		return errors.New("订单状态异常")
	}

	now := time.Now()
	order.Status = model.OrderStatusPaid
	order.PayChannel = model.PaymentChannelWechat
	order.BalancePaidAmount = 0
	order.WechatPaidAmount = order.TotalAmount
	order.PaymentOrderNo = orderNo
	order.PaidAt = &now
	if err := repository.UpdateOrder(order); err != nil {
		return errors.New("更新订单失败")
	}

	if err := handlePaidOrderPostActions(order.ID); err != nil {
		return err
	}

	return nil
}

func completeRechargePay(orderNo string) error {
	rechargeOrder, err := repository.GetRechargeOrderByOrderNo(orderNo)
	if err != nil {
		return errors.New("充值订单不存在")
	}
	if rechargeOrder.Status == model.RechargeOrderStatusPaid {
		return nil
	}
	if rechargeOrder.Status != model.RechargeOrderStatusPending {
		return errors.New("充值订单状态异常")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var order model.RechargeOrder
		if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == model.RechargeOrderStatusPaid {
			return nil
		}
		if order.Status != model.RechargeOrderStatusPending {
			return errors.New("充值订单状态异常")
		}

		balance, err := repository.GetUserBalanceByUserIDWithDB(tx, order.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			balance = &model.UserBalance{UserID: order.UserID}
			if err := tx.Create(balance).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		now := time.Now()
		balance.BalanceAmount += order.TotalArrivalAmount
		balance.TotalRechargeAmount += order.RechargeAmount
		balance.TotalGiftAmount += order.GiftAmount
		if err := repository.SaveUserBalanceWithDB(tx, balance); err != nil {
			return err
		}

		order.Status = model.RechargeOrderStatusPaid
		order.PayChannel = model.PaymentChannelWechat
		order.PaidAt = &now
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		orderID := order.ID
		activityID := order.ActivityID
		batch := &model.RechargeBatch{
			UserID:                  order.UserID,
			RechargeOrderID:         &orderID,
			ActivityID:              &activityID,
			ActivityNameSnapshot:    order.ActivityNameSnapshot,
			ActivityVersionSnapshot: normalizeRechargeActivityRuleVersion(order.ActivityVersionSnapshot),
			BatchType:               model.RechargeBatchTypeRecharge,
			PrincipalTotal:          order.RechargeAmount,
			GiftTotal:               order.GiftAmount,
			PrincipalRemaining:      order.RechargeAmount,
			GiftRemaining:           order.GiftAmount,
			PrincipalConsumeWeight:  normalizeConsumeWeight(order.PrincipalConsumeWeight, order.RechargeAmount),
			GiftConsumeWeight:       normalizeConsumeWeight(order.GiftConsumeWeight, order.GiftAmount),
		}
		if err := repository.CreateRechargeBatchWithDB(tx, batch); err != nil {
			return err
		}

		return repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
			UserID:                order.UserID,
			ChangeAmount:          order.TotalArrivalAmount,
			PrincipalChangeAmount: order.RechargeAmount,
			GiftChangeAmount:      order.GiftAmount,
			BalanceAfter:          balance.BalanceAmount,
			BizType:               model.BalanceTransactionTypeRecharge,
			SourceType:            model.BalanceTransactionSourceRechargeOrder,
			SourceID:              order.ID,
			RelatedBatchID:        &batch.ID,
			Remark:                fmt.Sprintf("充值%d元赠送%d元", order.RechargeAmount/100, order.GiftAmount/100),
		})
	})
}

func payOrderByBalance(userID, orderID uint) (*response.MiniProgramPayResp, error) {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权支付该订单")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("当前订单状态不可支付")
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		currentOrder, err := repository.GetOrderByIDWithDB(tx, orderID)
		if err != nil {
			return err
		}
		if currentOrder.Status != model.OrderStatusPending {
			return errors.New("当前订单状态不可支付")
		}

		balance, err := repository.GetUserBalanceByUserIDWithDB(tx, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("余额不足，请先充值")
		}
		if err != nil {
			return err
		}
		if balance.BalanceAmount < currentOrder.TotalAmount {
			return errors.New("余额不足，请先充值")
		}

		batches, err := repository.ListAvailableRechargeBatchesByUserIDWithDB(tx, userID)
		if err != nil {
			return err
		}
		trackedAmount := sumRechargeBatchAvailableAmount(batches)
		if trackedAmount < balance.BalanceAmount {
			openingBatch, openErr := repository.CreateLegacyOpeningRechargeBatchWithDB(tx, userID, balance.BalanceAmount-trackedAmount)
			if openErr != nil {
				return openErr
			}
			batches = append([]model.RechargeBatch{*openingBatch}, batches...)
		}

		remainingToConsume := currentOrder.TotalAmount
		principalConsumed := 0
		giftConsumed := 0
		allocationSeq := 1
		allocations := make([]model.BalanceConsumptionAllocation, 0)
		for i := range batches {
			batch := &batches[i]
			available := batch.PrincipalRemaining + batch.GiftRemaining
			if available <= 0 || remainingToConsume <= 0 {
				continue
			}
			consumeAmount := minInt(remainingToConsume, available)
			principalAmount, giftAmount := splitBatchConsumeAmount(batch.PrincipalRemaining, batch.GiftRemaining, consumeAmount)
			batch.PrincipalRemaining -= principalAmount
			batch.GiftRemaining -= giftAmount
			if err := repository.SaveRechargeBatchWithDB(tx, batch); err != nil {
				return err
			}
			allocations = append(allocations, model.BalanceConsumptionAllocation{
				UserID:          userID,
				OrderID:         currentOrder.ID,
				RechargeBatchID: batch.ID,
				AllocationSeq:   allocationSeq,
				TotalAmount:     consumeAmount,
				PrincipalAmount: principalAmount,
				GiftAmount:      giftAmount,
			})
			allocationSeq++
			principalConsumed += principalAmount
			giftConsumed += giftAmount
			remainingToConsume -= consumeAmount
		}
		if remainingToConsume > 0 {
			return errors.New("余额账户数据异常，请联系管理员")
		}
		if err := repository.CreateBalanceConsumptionAllocationsWithDB(tx, allocations); err != nil {
			return err
		}

		now := time.Now()
		balance.BalanceAmount -= currentOrder.TotalAmount
		balance.TotalConsumeAmount += currentOrder.TotalAmount
		if err := repository.SaveUserBalanceWithDB(tx, balance); err != nil {
			return err
		}

		currentOrder.Status = model.OrderStatusPaid
		currentOrder.PayChannel = model.PaymentChannelBalance
		currentOrder.BalancePaidAmount = currentOrder.TotalAmount
		currentOrder.WechatPaidAmount = 0
		currentOrder.PaymentOrderNo = ""
		currentOrder.PaidAt = &now
		if err := repository.UpdateOrderWithDB(tx, currentOrder); err != nil {
			return err
		}

		shopID := currentOrder.ShopID
		return repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
			UserID:                userID,
			ChangeAmount:          -currentOrder.TotalAmount,
			PrincipalChangeAmount: -principalConsumed,
			GiftChangeAmount:      -giftConsumed,
			BalanceAfter:          balance.BalanceAmount,
			BizType:               model.BalanceTransactionTypePay,
			SourceType:            model.BalanceTransactionSourceOrder,
			SourceID:              currentOrder.ID,
			ShopID:                &shopID,
			Remark:                "余额支付订单",
		})
	}); err != nil {
		return nil, err
	}

	if err := handlePaidOrderPostActions(orderID); err != nil {
		return nil, err
	}

	return &response.MiniProgramPayResp{Mode: payModeBalance}, nil
}

func buildPayBody(shopName string) string {
	shopName = strings.TrimSpace(shopName)
	if shopName == "" {
		return "DeskOrder订单"
	}

	body := "DeskOrder-" + shopName
	runes := []rune(body)
	if len(runes) > 64 {
		return string(runes[:64])
	}
	return body
}

func buildRechargePayBody(activityName string) string {
	activityName = strings.TrimSpace(activityName)
	if activityName == "" {
		return "DeskOrder余额充值"
	}
	body := "DeskOrder-余额充值-" + activityName
	runes := []rune(body)
	if len(runes) > 64 {
		return string(runes[:64])
	}
	return body
}

func normalizePayClientIP(clientIP string) string {
	parsedIP := net.ParseIP(strings.TrimSpace(clientIP))
	if parsedIP == nil {
		return "127.0.0.1"
	}
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		return ipv4.String()
	}
	return "127.0.0.1"
}

func createMockUserOrderPay(order *model.Order, mockResult string) (*response.MiniProgramPayResp, error) {
	if mockResult == mockPayResultSuccess {
		if err := PayNotify(order.OrderNo); err != nil {
			return nil, err
		}
	}

	return &response.MiniProgramPayResp{
		Mode:       payModeMock,
		MockResult: mockResult,
	}, nil
}

func createMockRechargePay(order *model.RechargeOrder, mockResult string) (*response.MiniProgramPayResp, error) {
	if mockResult == mockPayResultSuccess {
		if err := PayNotify(order.OrderNo); err != nil {
			return nil, err
		}
	}

	return &response.MiniProgramPayResp{
		Mode:       payModeMock,
		MockResult: mockResult,
	}, nil
}

func normalizeOrderPayMethod(method string) string {
	if strings.EqualFold(strings.TrimSpace(method), model.PaymentChannelBalance) {
		return model.PaymentChannelBalance
	}
	return model.PaymentChannelWechat
}

func handlePaidOrderPostActions(orderID uint) error {
	order, err := repository.GetOrderByID(orderID)
	if err != nil {
		return errors.New("查询订单失败")
	}
	if ws.DefaultHub != nil {
		ws.DefaultHub.PushToShop(order.ShopID, ws.Message{
			Type: ws.MsgTypeNewOrder,
			Data: order,
		})
	}
	if err := AutoAcceptOrder(order.ID); err != nil {
		log.Warn().Err(err).Uint("order_id", order.ID).Uint("shop_id", order.ShopID).Msg("auto accept order failed")
	}
	return nil
}

func sumRechargeBatchAvailableAmount(batches []model.RechargeBatch) int {
	total := 0
	for _, batch := range batches {
		total += batch.PrincipalRemaining + batch.GiftRemaining
	}
	return total
}

func splitBatchConsumeAmount(principalRemaining, giftRemaining, amount int) (int, int) {
	if amount <= 0 {
		return 0, 0
	}
	if giftRemaining <= 0 {
		return amount, 0
	}
	if principalRemaining <= 0 {
		return 0, amount
	}
	totalRemaining := principalRemaining + giftRemaining
	if totalRemaining <= 0 {
		return amount, 0
	}
	principalAmount := amount * principalRemaining / totalRemaining
	giftAmount := amount - principalAmount
	if principalAmount > principalRemaining {
		principalAmount = principalRemaining
		giftAmount = amount - principalAmount
	}
	if giftAmount > giftRemaining {
		giftAmount = giftRemaining
		principalAmount = amount - giftAmount
	}
	return principalAmount, giftAmount
}

func normalizeConsumeWeight(value, fallback int) int {
	if value > 0 {
		return value
	}
	if fallback > 0 {
		return fallback
	}
	return 0
}

func uniqueRechargeBatchIDs(allocations []model.BalanceConsumptionAllocation) []uint {
	seen := make(map[uint]struct{}, len(allocations))
	ids := make([]uint, 0, len(allocations))
	for _, allocation := range allocations {
		if _, ok := seen[allocation.RechargeBatchID]; ok {
			continue
		}
		seen[allocation.RechargeBatchID] = struct{}{}
		ids = append(ids, allocation.RechargeBatchID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func normalizePayMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), payModeReal) {
		return payModeReal
	}
	return payModeMock
}

func normalizeMockPayResult(result string) string {
	switch strings.ToLower(strings.TrimSpace(result)) {
	case mockPayResultCancel:
		return mockPayResultCancel
	case mockPayResultFail:
		return mockPayResultFail
	default:
		return mockPayResultSuccess
	}
}
