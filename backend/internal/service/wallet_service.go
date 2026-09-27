package service

import (
	"deskorder/config"
	"errors"
	"fmt"
	"strings"
	"time"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/database"
	"deskorder/pkg/wechat"

	"gorm.io/gorm"
)

const rechargeOrderExpireDuration = 30 * time.Minute

func GetUserProfile(userID uint) (*response.UserProfileResp, error) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	balance, err := repository.GetOrCreateUserBalanceByUserID(userID)
	if err != nil {
		return nil, errors.New("查询余额失败")
	}
	refundUnreadCount, err := repository.CountUnreadUserNotificationsByCategory(userID, model.UserNotificationCategoryRefund)
	if err != nil {
		return nil, errors.New("查询退款通知失败")
	}
	hasRefundRecordCount, err := repository.CountRefundRequestsByUserID(userID)
	if err != nil {
		return nil, errors.New("查询退款记录失败")
	}
	refundVisible, refundEnabled, err := GetMiniUserRefundFeatureConfig()
	if err != nil {
		return nil, errors.New("查询系统配置失败")
	}
	refundSubscribeTemplateIDs := GetRefundSubscribeTemplateIDs()

	return &response.UserProfileResp{
		ID:                         user.ID,
		Nickname:                   user.Nickname,
		Avatar:                     user.Avatar,
		Phone:                      user.Phone,
		BalanceAmount:              balance.BalanceAmount,
		TotalRechargeAmount:        balance.TotalRechargeAmount,
		TotalGiftAmount:            balance.TotalGiftAmount,
		TotalConsumeAmount:         balance.TotalConsumeAmount,
		RefundUnreadCount:          refundUnreadCount,
		HasRefundRecords:           hasRefundRecordCount > 0,
		RefundSubscribeEnabled:     len(refundSubscribeTemplateIDs) > 0,
		RefundSubscribeTemplateIDs: refundSubscribeTemplateIDs,
		RefundFeatureVisible:       refundVisible,
		RefundFeatureEnabled:       refundEnabled,
	}, nil
}

func UpdateUserProfile(userID uint, req *request.UpdateUserProfileReq) (*response.UserProfileResp, error) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	nickname := strings.TrimSpace(req.Nickname)
	avatar := strings.TrimSpace(req.Avatar)
	if nickname == "" && avatar == "" {
		return nil, errors.New("请先提供昵称或头像")
	}
	changed := false
	if nickname != "" && nickname != user.Nickname {
		user.Nickname = nickname
		changed = true
	}
	if avatar != "" && avatar != user.Avatar {
		user.Avatar = avatar
		changed = true
	}
	if changed {
		if err := repository.UpdateUser(user); err != nil {
			return nil, errors.New("更新用户资料失败")
		}
	}
	return GetUserProfile(userID)
}

func BindUserPhone(userID uint, req *request.BindUserPhoneReq) (*response.UserProfileResp, error) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	cfg := config.AppConfig.WeChat
	phoneResp, err := wechat.GetUserPhoneNumber(cfg.UserAppID, cfg.UserSecret, strings.TrimSpace(req.Code))
	if err != nil {
		return nil, errors.New("微信手机号获取失败")
	}
	phone := strings.TrimSpace(phoneResp.PhoneInfo.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(phoneResp.PhoneInfo.PhoneNumber)
	}
	if phone == "" {
		return nil, errors.New("未获取到有效手机号")
	}
	if user.Phone != phone {
		user.Phone = phone
		if err := repository.UpdateUser(user); err != nil {
			return nil, errors.New("更新用户手机号失败")
		}
	}
	return GetUserProfile(userID)
}

func GetRechargeActivities() ([]response.RechargeActivityResp, error) {
	activities, err := repository.GetActiveRechargeActivities(time.Now())
	if err != nil {
		return nil, errors.New("查询充值活动失败")
	}
	list := make([]response.RechargeActivityResp, 0, len(activities))
	for _, activity := range activities {
		list = append(list, rechargeActivityToResp(&activity))
	}
	return list, nil
}

func CreateRechargeOrder(userID uint, req *request.CreateRechargeOrderReq) (*response.RechargeOrderResp, error) {
	activity, err := repository.GetRechargeActivityByID(req.ActivityID)
	if err != nil {
		return nil, errors.New("充值活动不存在")
	}
	if !isRechargeActivityAvailable(activity, time.Now()) {
		return nil, errors.New("充值活动不可用")
	}
	now := time.Now()
	expiresAt := now.Add(rechargeOrderExpireDuration)
	ruleVersion := normalizeRechargeActivityRuleVersion(activity.RuleVersion)

	order := &model.RechargeOrder{
		OrderNo:                 fmt.Sprintf("RO%d%d", time.Now().UnixNano(), userID),
		UserID:                  userID,
		ActivityID:              activity.ID,
		ActivityNameSnapshot:    activity.Name,
		ActivityVersionSnapshot: ruleVersion,
		RechargeAmount:          activity.RechargeAmount,
		GiftAmount:              activity.GiftAmount,
		PrincipalConsumeWeight:  activity.RechargeAmount,
		GiftConsumeWeight:       activity.GiftAmount,
		TotalArrivalAmount:      activity.RechargeAmount + activity.GiftAmount,
		PayAmount:               activity.RechargeAmount,
		PayChannel:              model.PaymentChannelWechat,
		Status:                  model.RechargeOrderStatusPending,
		ExpiresAt:               &expiresAt,
	}
	if err := repository.CreateRechargeOrder(order); err != nil {
		return nil, errors.New("创建充值订单失败")
	}
	created, err := repository.GetRechargeOrderByID(order.ID)
	if err != nil {
		return nil, errors.New("查询充值订单失败")
	}
	refundVisible, refundEnabled, _ := GetMiniUserRefundFeatureConfig()
	return rechargeOrderToResp(created, model.RefundRequest{}, refundVisible, refundEnabled), nil
}

func GetUserRechargeOrders(userID uint, page, size int, status *int) ([]response.RechargeOrderResp, int64, error) {
	list, total, err := repository.GetRechargeOrdersByUserID(userID, page, size, status)
	if err != nil {
		return nil, 0, errors.New("查询充值订单失败")
	}
	refundVisible, refundEnabled, _ := GetMiniUserRefundFeatureConfig()
	orderIDs := make([]uint, 0, len(list))
	for i := range list {
		orderIDs = append(orderIDs, list[i].ID)
	}
	refundMap, err := repository.ListLatestRefundRequestsBySourceIDs(model.RefundSourceTypeRecharge, orderIDs)
	if err != nil {
		return nil, 0, errors.New("查询充值退款状态失败")
	}
	respList := make([]response.RechargeOrderResp, 0, len(list))
	for i := range list {
		respList = append(respList, *rechargeOrderToResp(&list[i], refundMap[list[i].ID], refundVisible, refundEnabled))
	}
	return respList, total, nil
}

func CancelUserRechargeOrder(userID, orderID uint) (*response.RechargeOrderResp, error) {
	order, err := repository.GetRechargeOrderByUserIDAndID(userID, orderID)
	if err != nil {
		return nil, errors.New("充值订单不存在")
	}
	if order.Status != model.RechargeOrderStatusPending {
		return nil, errors.New("当前充值订单状态不可取消")
	}
	order.Status = model.RechargeOrderStatusCancelled
	if err := repository.UpdateRechargeOrder(order); err != nil {
		return nil, errors.New("取消充值订单失败")
	}
	refundVisible, refundEnabled, _ := GetMiniUserRefundFeatureConfig()
	return rechargeOrderToResp(order, model.RefundRequest{}, refundVisible, refundEnabled), nil
}

func ApplyUserRechargeRefund(userID, rechargeOrderID uint, requestNote string) (*response.RechargeOrderResp, error) {
	refundVisible, refundEnabled, err := GetMiniUserRefundFeatureConfig()
	if err != nil {
		return nil, errors.New("查询系统配置失败")
	}
	if !refundVisible || !refundEnabled {
		return nil, errors.New("当前暂不支持自助充值退款申请")
	}

	ctx, err := loadRechargeRefundContext(rechargeOrderID)
	if err != nil {
		return nil, err
	}
	if ctx.order.UserID != userID {
		return nil, errors.New("无权操作该充值订单")
	}
	if ctx.batch.PrincipalRemaining <= 0 {
		return nil, errors.New("当前充值订单无可退本金")
	}
	if existing, findErr := repository.GetLatestOpenRefundRequestBySource(model.RefundSourceTypeRecharge, ctx.order.ID); findErr == nil && existing != nil {
		if existing.Status == model.RefundStatusPendingReview || existing.Status == model.RefundStatusApproved || existing.Status == model.RefundStatusProcessing {
			return nil, errors.New("当前充值订单已有待处理退款申请")
		}
		if existing.Status == model.RefundStatusSuccess {
			return nil, errors.New("当前充值订单已有成功退款记录")
		}
	}

	var refundItem *model.RefundRequest
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		refundItem, err = createRechargeRefundRequestWithDB(tx, ctx, model.RefundRequestChannelUser, strings.TrimSpace(requestNote))
		return err
	}); err != nil {
		return nil, errors.New("提交充值退款申请失败")
	}

	order, err := repository.GetRechargeOrderByUserIDAndID(userID, rechargeOrderID)
	if err != nil {
		return nil, errors.New("充值订单不存在")
	}
	return rechargeOrderToResp(order, *refundItem, refundVisible, refundEnabled), nil
}

func GetBalanceTransactions(userID uint, page, size int) ([]response.BalanceTransactionResp, int64, error) {
	list, total, err := repository.GetBalanceTransactionsByUserID(userID, page, size)
	if err != nil {
		return nil, 0, errors.New("查询余额流水失败")
	}
	respList := make([]response.BalanceTransactionResp, 0, len(list))
	for _, item := range list {
		respList = append(respList, balanceTransactionToResp(&item))
	}
	return respList, total, nil
}

func rechargeActivityToResp(activity *model.RechargeActivity) response.RechargeActivityResp {
	return response.RechargeActivityResp{
		ID:             activity.ID,
		Name:           activity.Name,
		RechargeAmount: activity.RechargeAmount,
		GiftAmount:     activity.GiftAmount,
		ArrivalAmount:  activity.RechargeAmount + activity.GiftAmount,
		Status:         activity.Status,
		Sort:           activity.Sort,
		StartAt:        activity.StartAt,
		EndAt:          activity.EndAt,
		Description:    activity.Description,
	}
}

func rechargeOrderToResp(order *model.RechargeOrder, refund model.RefundRequest, refundVisible, refundEnabled bool) *response.RechargeOrderResp {
	canApplyRefund := canUserApplyRechargeRefund(order, &refund, refundVisible, refundEnabled)
	refundActionText := ""
	if canApplyRefund {
		refundActionText = "申请退款"
	}
	resp := &response.RechargeOrderResp{
		ID:                 order.ID,
		OrderNo:            order.OrderNo,
		ActivityID:         order.ActivityID,
		ActivityName:       order.ActivityNameSnapshot,
		RechargeAmount:     order.RechargeAmount,
		GiftAmount:         order.GiftAmount,
		TotalArrivalAmount: order.TotalArrivalAmount,
		PayAmount:          order.PayAmount,
		PayChannel:         order.PayChannel,
		Status:             order.Status,
		CanApplyRefund:     canApplyRefund,
		RefundEntryVisible: refundVisible,
		RefundActionText:   refundActionText,
		PaidAt:             order.PaidAt,
		CreatedAt:          order.CreatedAt,
	}
	if refund.ID > 0 {
		resp.RefundStatus = refund.Status
		resp.RefundStatusText = userRechargeRefundStatusText(refund.Status)
		resp.RefundResultText = buildUserRechargeRefundResultText(&refund)
		if refund.ReviewedAt != nil {
			resp.RefundUpdatedAt = refund.ReviewedAt
		} else {
			updatedAt := refund.UpdatedAt
			resp.RefundUpdatedAt = &updatedAt
		}
	}
	return resp
}

func canUserApplyRechargeRefund(order *model.RechargeOrder, refund *model.RefundRequest, refundVisible, refundEnabled bool) bool {
	if order == nil || !refundVisible || !refundEnabled || order.Status != model.RechargeOrderStatusPaid {
		return false
	}
	if refund != nil && refund.ID > 0 {
		switch refund.Status {
		case model.RefundStatusPendingReview, model.RefundStatusApproved, model.RefundStatusProcessing, model.RefundStatusSuccess:
			return false
		}
	}
	batch, err := repository.GetRechargeBatchByRechargeOrderID(order.ID)
	if err != nil {
		return false
	}
	return batch.PrincipalRemaining > 0
}

func balanceTransactionToResp(record *model.BalanceTransaction) response.BalanceTransactionResp {
	resp := response.BalanceTransactionResp{
		ID:           record.ID,
		ChangeAmount: record.ChangeAmount,
		BalanceAfter: record.BalanceAfter,
		BizType:      record.BizType,
		SourceType:   record.SourceType,
		SourceID:     record.SourceID,
		ShopID:       record.ShopID,
		Remark:       record.Remark,
		CreatedAt:    record.CreatedAt,
	}
	if record.ShopID != nil {
		resp.ShopName = record.Shop.Name
	}
	return resp
}

func isRechargeActivityAvailable(activity *model.RechargeActivity, now time.Time) bool {
	if activity == nil || activity.Status != model.RechargeActivityStatusEnabled {
		return false
	}
	if activity.StartAt != nil && activity.StartAt.After(now) {
		return false
	}
	if activity.EndAt != nil && activity.EndAt.Before(now) {
		return false
	}
	return true
}

func normalizeRechargeActivityRuleVersion(version int) int {
	if version < 1 {
		return 1
	}
	return version
}

func userRechargeRefundStatusText(status string) string {
	switch status {
	case model.RefundStatusPendingReview:
		return "退款待审核"
	case model.RefundStatusCancelled:
		return "退款已撤回"
	case model.RefundStatusProcessing:
		return "退款处理中"
	case model.RefundStatusSuccess:
		return "已退款"
	case model.RefundStatusRejected:
		return "退款已驳回"
	case model.RefundStatusFailed:
		return "退款失败"
	default:
		return ""
	}
}

func buildUserRechargeRefundResultText(item *model.RefundRequest) string {
	if item == nil {
		return ""
	}
	switch item.Status {
	case model.RefundStatusPendingReview:
		return "平台正在审核你的充值退款请求，请稍后查看处理结果。"
	case model.RefundStatusCancelled:
		return "你已撤回本次充值退款申请，如仍需退款可重新提交。"
	case model.RefundStatusProcessing:
		return "平台已受理退款，正在按原支付路径处理。"
	case model.RefundStatusSuccess:
		return "充值退款已处理完成，请留意原支付账户到账情况。"
	case model.RefundStatusRejected:
		if item.RejectReason != "" {
			return item.RejectReason
		}
		return "本次充值退款未通过审核，如有疑问请联系平台。"
	case model.RefundStatusFailed:
		return "本次充值退款暂未成功，平台会继续跟进处理。"
	default:
		return ""
	}
}
