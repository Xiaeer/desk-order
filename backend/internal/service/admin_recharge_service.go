package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

var (
	errLegacyRechargeOrderAmbiguous       = errors.New("该用户存在多笔未建账旧充值，无法按订单精确清理，请先人工处理历史余额")
	errLegacyRechargeOrderNoRemaining     = errors.New("旧充值订单未建批次，当前无可清理历史余额")
	errLegacyRechargeOrderBalanceMismatch = errors.New("旧充值订单历史余额与订单金额不匹配，请先人工处理")
)

func AdminGetRechargeActivities(page, size int, status *int) ([]response.AdminRechargeActivityResp, int64, error) {
	activities, total, err := repository.GetRechargeActivities(page, size, status)
	if err != nil {
		return nil, 0, errors.New("获取充值活动失败")
	}
	list := make([]response.AdminRechargeActivityResp, 0, len(activities))
	for _, activity := range activities {
		list = append(list, adminRechargeActivityToResp(&activity))
	}
	return list, total, nil
}

func AdminGetRechargeOrders(page, size int, status *int, orderNo, userKeyword, createdStartAtRaw, createdEndAtRaw, suggestedAction string) ([]response.AdminRechargeOrderResp, int64, error) {
	if err := validateAdminRechargeOrderStatus(status); err != nil {
		return nil, 0, err
	}
	if err := validateAdminRechargeOrderSuggestedAction(suggestedAction); err != nil {
		return nil, 0, err
	}
	createdStartAt, createdEndAt, err := parseAdminRechargeOrderTimeRange(createdStartAtRaw, createdEndAtRaw)
	if err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(suggestedAction) != "" {
		return adminGetRechargeOrdersWithDerivedFilters(page, size, status, orderNo, userKeyword, createdStartAt, createdEndAt, suggestedAction)
	}
	orders, total, err := repository.GetAdminRechargeOrders(page, size, status, orderNo, userKeyword, createdStartAt, createdEndAt)
	if err != nil {
		return nil, 0, errors.New("获取充值订单失败")
	}
	orderIDs := make([]uint, 0, len(orders))
	for i := range orders {
		orderIDs = append(orderIDs, orders[i].ID)
	}
	refundMap, err := repository.ListLatestRefundRequestsBySourceIDs(model.RefundSourceTypeRecharge, orderIDs)
	if err != nil {
		return nil, 0, errors.New("获取充值订单退款状态失败")
	}
	actualBatchMap, err := listAdminRechargeOrderActualBatchMap(orderIDs)
	if err != nil {
		return nil, 0, errors.New("获取充值订单批次状态失败")
	}
	displayBatchMap, err := listAdminRechargeOrderListDisplayBatchMap(orders, actualBatchMap)
	if err != nil {
		return nil, 0, errors.New("获取充值订单剩余余额失败")
	}
	list := make([]response.AdminRechargeOrderResp, 0, len(orders))
	for i := range orders {
		item := adminRechargeOrderToResp(&orders[i], refundMap[orders[i].ID], displayBatchMap[orders[i].ID])
		applyAdminRechargeRefundState(&item, &orders[i], refundMap[orders[i].ID], actualBatchMap[orders[i].ID])
		applyAdminRechargeOrderListState(&item, &orders[i], refundMap[orders[i].ID], displayBatchMap[orders[i].ID])
		list = append(list, item)
	}
	return list, total, nil
}

func adminGetRechargeOrdersWithDerivedFilters(page, size int, status *int, orderNo, userKeyword string, createdStartAt, createdEndAt *time.Time, suggestedAction string) ([]response.AdminRechargeOrderResp, int64, error) {
	orders, err := repository.ListAdminRechargeOrders(status, orderNo, userKeyword, createdStartAt, createdEndAt)
	if err != nil {
		return nil, 0, errors.New("获取充值订单失败")
	}
	orderIDs := make([]uint, 0, len(orders))
	for i := range orders {
		orderIDs = append(orderIDs, orders[i].ID)
	}
	refundMap, err := repository.ListLatestRefundRequestsBySourceIDs(model.RefundSourceTypeRecharge, orderIDs)
	if err != nil {
		return nil, 0, errors.New("获取充值订单退款状态失败")
	}
	actualBatchMap, err := listAdminRechargeOrderActualBatchMap(orderIDs)
	if err != nil {
		return nil, 0, errors.New("获取充值订单批次状态失败")
	}
	displayBatchMap, err := listAdminRechargeOrderListDisplayBatchMap(orders, actualBatchMap)
	if err != nil {
		return nil, 0, errors.New("获取充值订单剩余余额失败")
	}
	matched := make([]response.AdminRechargeOrderResp, 0, len(orders))
	for i := range orders {
		item := adminRechargeOrderToResp(&orders[i], refundMap[orders[i].ID], displayBatchMap[orders[i].ID])
		applyAdminRechargeRefundState(&item, &orders[i], refundMap[orders[i].ID], actualBatchMap[orders[i].ID])
		applyAdminRechargeOrderListState(&item, &orders[i], refundMap[orders[i].ID], displayBatchMap[orders[i].ID])
		if !matchesAdminRechargeOrderDerivedFilters(&item, suggestedAction) {
			continue
		}
		matched = append(matched, item)
	}
	total := int64(len(matched))
	if total == 0 {
		return []response.AdminRechargeOrderResp{}, 0, nil
	}
	start := (page - 1) * size
	if start >= len(matched) {
		return []response.AdminRechargeOrderResp{}, total, nil
	}
	end := start + size
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}

func AdminGetRechargeOrderDetail(id uint) (*response.AdminRechargeOrderResp, error) {
	order, err := repository.GetAdminRechargeOrderByID(id)
	if err != nil {
		return nil, errors.New("充值订单不存在")
	}
	refundMap, err := repository.ListLatestRefundRequestsBySourceIDs(model.RefundSourceTypeRecharge, []uint{order.ID})
	if err != nil {
		return nil, errors.New("获取充值订单退款状态失败")
	}
	actualBatch, err := resolveAdminRechargeOrderRefundBatch(order.ID)
	if err != nil {
		return nil, errors.New("获取充值订单退款状态失败")
	}
	batch, cleanupUnavailableReason, err := resolveAdminRechargeOrderDisplayBatch(order)
	if err != nil {
		return nil, errors.New("获取充值订单剩余余额失败")
	}
	resp := adminRechargeOrderToResp(order, refundMap[order.ID], batch)
	applyAdminRechargeCleanupState(&resp, order, refundMap[order.ID], batch, cleanupUnavailableReason)
	applyAdminRechargeRefundState(&resp, order, refundMap[order.ID], actualBatch)
	return &resp, nil
}

func AdminCleanupRechargeOrderRemaining(adminID, rechargeOrderID uint, req *request.AdminCleanupRechargeOrderReq) (*response.AdminRechargeOrderResp, error) {
	cleanupRemark := strings.TrimSpace(req.Remark)
	if cleanupRemark == "" {
		cleanupRemark = "作废此批次剩余"
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		order, err := repository.GetRechargeOrderByIDWithDB(tx, rechargeOrderID)
		if err != nil {
			return errors.New("充值订单不存在")
		}
		if order.Status != model.RechargeOrderStatusPaid {
			return errors.New("当前充值订单状态不可清理")
		}
		batch, err := resolveAdminRechargeOrderCleanupBatch(tx, order)
		if err != nil {
			return err
		}
		principalAmount := batch.PrincipalRemaining
		giftAmount := batch.GiftRemaining
		totalAmount := principalAmount + giftAmount
		if totalAmount <= 0 {
			return errors.New("当前充值订单无可清理剩余余额")
		}
		existingRefund, refundErr := repository.GetLatestOpenRefundRequestBySourceWithDB(tx, model.RefundSourceTypeRecharge, order.ID)
		if refundErr == nil && existingRefund != nil {
			return errors.New("当前充值订单已有待处理退款申请")
		}
		if refundErr != nil && !errors.Is(refundErr, gorm.ErrRecordNotFound) {
			return refundErr
		}
		balance, err := repository.GetUserBalanceByUserIDWithDB(tx, order.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户余额不存在")
		}
		if err != nil {
			return err
		}
		if balance.BalanceAmount < totalAmount {
			return errors.New("用户当前余额不足，无法作废此批次剩余，请人工处理")
		}

		balance.BalanceAmount -= totalAmount
		if balance.TotalRechargeAmount >= principalAmount {
			balance.TotalRechargeAmount -= principalAmount
		} else {
			balance.TotalRechargeAmount = 0
		}
		if balance.TotalGiftAmount >= giftAmount {
			balance.TotalGiftAmount -= giftAmount
		} else {
			balance.TotalGiftAmount = 0
		}
		if err := repository.SaveUserBalanceWithDB(tx, balance); err != nil {
			return err
		}

		batch.PrincipalRemaining = 0
		batch.GiftRemaining = 0
		if err := repository.SaveRechargeBatchWithDB(tx, batch); err != nil {
			return err
		}

		remark := buildAdminRechargeCleanupRemark(adminID, order.OrderNo, cleanupRemark)
		return repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
			UserID:                order.UserID,
			ChangeAmount:          -totalAmount,
			PrincipalChangeAmount: -principalAmount,
			GiftChangeAmount:      -giftAmount,
			BalanceAfter:          balance.BalanceAmount,
			BizType:               model.BalanceTransactionTypeAdjust,
			SourceType:            model.BalanceTransactionSourceRechargeOrder,
			SourceID:              order.ID,
			RelatedBatchID:        &batch.ID,
			Remark:                remark,
		})
	})
	if err != nil {
		switch err.Error() {
		case "充值订单不存在", "充值批次不存在", "用户余额不存在", "当前充值订单状态不可清理", "当前充值订单无可清理剩余余额", "当前充值订单已有待处理退款申请", "用户当前余额不足，无法作废此批次剩余，请人工处理", errLegacyRechargeOrderAmbiguous.Error(), errLegacyRechargeOrderNoRemaining.Error(), errLegacyRechargeOrderBalanceMismatch.Error():
			return nil, err
		default:
			return nil, errors.New("作废此批次剩余失败")
		}
	}
	return AdminGetRechargeOrderDetail(rechargeOrderID)
}

func AdminCreateRechargeActivity(req *request.AdminRechargeActivityReq) (*response.AdminRechargeActivityResp, error) {
	status, err := normalizeAdminRechargeActivityStatus(req.Status)
	if err != nil {
		return nil, err
	}
	if err := validateRechargeActivityWindow(req.StartAt, req.EndAt); err != nil {
		return nil, err
	}
	activity := &model.RechargeActivity{
		Name:           req.Name,
		RechargeAmount: req.RechargeAmount,
		GiftAmount:     req.GiftAmount,
		RuleVersion:    1,
		Status:         status,
		Sort:           req.Sort,
		StartAt:        req.StartAt,
		EndAt:          req.EndAt,
		Description:    req.Description,
	}
	if err := repository.CreateRechargeActivity(activity); err != nil {
		return nil, errors.New("创建充值活动失败")
	}
	returnResponse := adminRechargeActivityToResp(activity)
	return &returnResponse, nil
}

func AdminUpdateRechargeActivity(id uint, req *request.AdminRechargeActivityReq) (*response.AdminRechargeActivityResp, error) {
	status, err := normalizeAdminRechargeActivityStatus(req.Status)
	if err != nil {
		return nil, err
	}
	if err := validateRechargeActivityWindow(req.StartAt, req.EndAt); err != nil {
		return nil, err
	}
	activity, err := repository.GetRechargeActivityByID(id)
	if err != nil {
		return nil, errors.New("充值活动不存在")
	}
	baseVersion := normalizeRechargeActivityRuleVersion(activity.RuleVersion)
	if shouldBumpRechargeActivityRuleVersion(activity, req) {
		activity.RuleVersion = baseVersion + 1
	} else {
		activity.RuleVersion = baseVersion
	}
	activity.Name = req.Name
	activity.RechargeAmount = req.RechargeAmount
	activity.GiftAmount = req.GiftAmount
	activity.Status = status
	activity.Sort = req.Sort
	activity.StartAt = req.StartAt
	activity.EndAt = req.EndAt
	activity.Description = req.Description
	if err := repository.UpdateRechargeActivity(activity); err != nil {
		return nil, errors.New("更新充值活动失败")
	}
	resp := adminRechargeActivityToResp(activity)
	return &resp, nil
}

func AdminDeleteRechargeActivity(id uint) error {
	activity, err := repository.GetRechargeActivityByID(id)
	if err != nil {
		return errors.New("充值活动不存在")
	}
	if activity.Status != model.RechargeActivityStatusDisabled {
		return errors.New("仅允许删除已下线活动")
	}
	orderCount, err := repository.CountRechargeOrdersByActivityID(id)
	if err != nil {
		return errors.New("检查充值活动关联订单失败")
	}
	if orderCount > 0 {
		return errors.New("充值活动已有关联充值订单，不能删除")
	}
	if err := repository.DeleteRechargeActivity(id); err != nil {
		return errors.New("删除充值活动失败")
	}
	return nil
}

func validateRechargeActivityWindow(startAt, endAt *time.Time) error {
	if startAt != nil && endAt != nil && startAt.After(*endAt) {
		return errors.New("活动开始时间不能晚于结束时间")
	}
	return nil
}

func normalizeAdminRechargeActivityStatus(status *int) (int, error) {
	if status == nil {
		return 0, errors.New("充值活动状态不能为空")
	}
	if *status != model.RechargeActivityStatusDisabled && *status != model.RechargeActivityStatusEnabled {
		return 0, errors.New("充值活动状态无效")
	}
	return *status, nil
}

func shouldBumpRechargeActivityRuleVersion(activity *model.RechargeActivity, req *request.AdminRechargeActivityReq) bool {
	if activity == nil || req == nil {
		return false
	}
	return activity.RechargeAmount != req.RechargeAmount || activity.GiftAmount != req.GiftAmount
}

func adminRechargeActivityToResp(activity *model.RechargeActivity) response.AdminRechargeActivityResp {
	return response.AdminRechargeActivityResp{
		ID:             activity.ID,
		Name:           activity.Name,
		RechargeAmount: activity.RechargeAmount,
		GiftAmount:     activity.GiftAmount,
		ArrivalAmount:  activity.RechargeAmount + activity.GiftAmount,
		Status:         activity.Status,
		Sort:           activity.Sort,
		StartAt:        formatAdminTime(activity.StartAt),
		EndAt:          formatAdminTime(activity.EndAt),
		Description:    activity.Description,
		CreatedAt:      activity.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      activity.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func adminRechargeOrderToResp(order *model.RechargeOrder, refund model.RefundRequest, batch *model.RechargeBatch) response.AdminRechargeOrderResp {
	activityName := strings.TrimSpace(order.ActivityNameSnapshot)
	if activityName == "" {
		activityName = order.Activity.Name
	}
	resp := response.AdminRechargeOrderResp{
		ID:                       order.ID,
		OrderNo:                  order.OrderNo,
		UserID:                   order.UserID,
		UserDisplayName:          buildAdminUserDisplayName(&order.User),
		UserIdentityLabel:        buildAdminUserIdentityLabel(&order.User),
		UserNickname:             order.User.Nickname,
		UserPhone:                order.User.Phone,
		ActivityID:               order.ActivityID,
		ActivityName:             activityName,
		ActivityVersion:          order.ActivityVersionSnapshot,
		RechargeAmount:           order.RechargeAmount,
		GiftAmount:               order.GiftAmount,
		TotalArrivalAmount:       order.TotalArrivalAmount,
		PayAmount:                order.PayAmount,
		PayChannel:               order.PayChannel,
		PayChannelText:           adminRechargeOrderPayChannelText(order.PayChannel),
		Status:                   order.Status,
		StatusText:               adminRechargeOrderStatusText(order.Status),
		ExpiresAt:                formatAdminTime(order.ExpiresAt),
		PaidAt:                   formatAdminTime(order.PaidAt),
		CreatedAt:                order.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:                order.UpdatedAt.Format("2006-01-02 15:04:05"),
		RefundRequestID:          refund.ID,
		RefundRequestNo:          refund.RequestNo,
		RefundStatus:             refund.Status,
		RefundStatusText:         adminRefundStatusText(refund.Status),
		RefundRequestedAmount:    refund.RequestedTotalAmount,
		RefundApprovedAmount:     refund.ApprovedTotalAmount,
		RefundRequestChannel:     refund.RequestChannel,
		RefundRequestChannelText: adminRefundRequestChannelText(refund.RequestChannel),
		RefundUpdatedAt:          formatAdminTime(&refund.UpdatedAt),
	}
	if batch != nil {
		resp.CurrentPrincipalRemaining = batch.PrincipalRemaining
		resp.CurrentGiftRemaining = batch.GiftRemaining
		resp.CurrentRemainingAmount = batch.PrincipalRemaining + batch.GiftRemaining
		resp.HasRemainingSnapshot = true
		resp.CanCleanupRemaining = order.Status == model.RechargeOrderStatusPaid && resp.CurrentRemainingAmount > 0 && !isBlockingRechargeRefundStatus(refund.Status)
	}
	return resp
}

func matchesAdminRechargeOrderDerivedFilters(item *response.AdminRechargeOrderResp, suggestedAction string) bool {
	if item == nil {
		return false
	}
	if trimmedSuggestedAction := strings.TrimSpace(suggestedAction); trimmedSuggestedAction != "" && item.SuggestedAction != trimmedSuggestedAction {
		return false
	}
	return true
}

func listAdminRechargeOrderActualBatchMap(orderIDs []uint) (map[uint]*model.RechargeBatch, error) {
	result := make(map[uint]*model.RechargeBatch, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}
	batches, err := repository.ListRechargeBatchesByRechargeOrderIDs(orderIDs)
	if err != nil {
		return nil, err
	}
	for i := range batches {
		batch := batches[i]
		if batch.RechargeOrderID == nil {
			continue
		}
		result[*batch.RechargeOrderID] = &batch
	}
	return result, nil
}

func listAdminRechargeOrderListDisplayBatchMap(orders []model.RechargeOrder, actualBatchMap map[uint]*model.RechargeBatch) (map[uint]*model.RechargeBatch, error) {
	result := make(map[uint]*model.RechargeBatch, len(orders))
	for i := range orders {
		order := &orders[i]
		if batch := actualBatchMap[order.ID]; batch != nil {
			result[order.ID] = batch
			continue
		}
		if order.Status != model.RechargeOrderStatusPaid {
			continue
		}
		displayBatch, _, err := resolveAdminRechargeOrderDisplayBatch(order)
		if err != nil {
			return nil, err
		}
		if displayBatch == nil {
			continue
		}
		if displayBatch.PrincipalRemaining+displayBatch.GiftRemaining > 0 {
			continue
		}
		result[order.ID] = displayBatch
	}
	return result, nil
}

func resolveAdminRechargeOrderDisplayBatch(order *model.RechargeOrder) (*model.RechargeBatch, string, error) {
	batch, err := repository.GetRechargeBatchByRechargeOrderID(order.ID)
	if err == nil {
		return batch, "", nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}
	plan, err := planLegacyRechargeOrderBatch(database.DB, order)
	if err != nil {
		if errors.Is(err, errLegacyRechargeOrderNoRemaining) && order != nil && order.Status == model.RechargeOrderStatusPaid {
			return buildLegacyRechargeOrderBatchPreview(order, 0), "", nil
		}
		if isLegacyRechargeOrderPlanError(err) {
			return nil, err.Error(), nil
		}
		return nil, "", err
	}
	return plan.batch, "", nil
}

func resolveAdminRechargeOrderRefundBatch(rechargeOrderID uint) (*model.RechargeBatch, error) {
	batch, err := repository.GetRechargeBatchByRechargeOrderID(rechargeOrderID)
	if err == nil {
		return batch, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return nil, err
}

func resolveAdminRechargeOrderCleanupBatch(tx *gorm.DB, order *model.RechargeOrder) (*model.RechargeBatch, error) {
	batch, err := repository.GetRechargeBatchByRechargeOrderIDWithDB(tx, order.ID)
	if err == nil {
		return batch, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	plan, err := planLegacyRechargeOrderBatch(tx, order)
	if err != nil {
		return nil, err
	}
	if !plan.shouldCreate {
		return plan.batch, nil
	}
	if err := repository.CreateRechargeBatchWithDB(tx, plan.batch); err != nil {
		return nil, err
	}
	return plan.batch, nil
}

type legacyRechargeOrderBatchPlan struct {
	batch        *model.RechargeBatch
	shouldCreate bool
}

func planLegacyRechargeOrderBatch(db *gorm.DB, order *model.RechargeOrder) (*legacyRechargeOrderBatchPlan, error) {
	if order == nil || order.Status != model.RechargeOrderStatusPaid {
		return nil, errLegacyRechargeOrderNoRemaining
	}
	availableBatches, err := listAdminRechargeOrderAvailableBatches(db, order.UserID)
	if err != nil {
		return nil, err
	}
	openingBatch, err := findResolvableLegacyOpeningBatch(availableBatches)
	if err != nil {
		return nil, err
	}
	remainingAmount := 0
	if openingBatch == nil {
		balance, err := repository.GetUserBalanceByUserIDWithDB(db, order.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errLegacyRechargeOrderNoRemaining
		}
		if err != nil {
			return nil, err
		}
		remainingAmount = balance.BalanceAmount - sumRechargeBatchAvailableAmount(availableBatches)
		if remainingAmount <= 0 {
			return nil, errLegacyRechargeOrderNoRemaining
		}
	}
	missingOrders, err := repository.ListPaidRechargeOrdersWithoutBatchByUserIDWithDB(db, order.UserID)
	if err != nil {
		return nil, err
	}
	if len(missingOrders) != 1 || missingOrders[0].ID != order.ID {
		return nil, errLegacyRechargeOrderAmbiguous
	}
	if openingBatch != nil {
		if openingBatch.PrincipalRemaining+openingBatch.GiftRemaining > order.TotalArrivalAmount {
			return nil, errLegacyRechargeOrderBalanceMismatch
		}
		return &legacyRechargeOrderBatchPlan{batch: openingBatch}, nil
	}
	if remainingAmount > order.TotalArrivalAmount {
		return nil, errLegacyRechargeOrderBalanceMismatch
	}
	return &legacyRechargeOrderBatchPlan{
		batch:        buildLegacyRechargeOrderBatchPreview(order, remainingAmount),
		shouldCreate: true,
	}, nil
}

func listAdminRechargeOrderAvailableBatches(db *gorm.DB, userID uint) ([]model.RechargeBatch, error) {
	if db == database.DB {
		return repository.ListAvailableRechargeBatchesByUserID(userID)
	}
	return repository.ListAvailableRechargeBatchesByUserIDWithDB(db, userID)
}

func findResolvableLegacyOpeningBatch(batches []model.RechargeBatch) (*model.RechargeBatch, error) {
	var openingBatch *model.RechargeBatch
	for i := range batches {
		batch := &batches[i]
		if batch.BatchType != model.RechargeBatchTypeOpening {
			continue
		}
		if batch.PrincipalRemaining+batch.GiftRemaining <= 0 {
			continue
		}
		if openingBatch != nil {
			return nil, errLegacyRechargeOrderAmbiguous
		}
		copyBatch := *batch
		openingBatch = &copyBatch
	}
	return openingBatch, nil
}

func buildLegacyRechargeOrderBatchPreview(order *model.RechargeOrder, remainingAmount int) *model.RechargeBatch {
	principalRemaining, giftRemaining := splitBatchConsumeAmount(order.RechargeAmount, order.GiftAmount, remainingAmount)
	orderID := order.ID
	activityID := order.ActivityID
	activityName := strings.TrimSpace(order.ActivityNameSnapshot)
	if activityName == "" {
		activityName = order.Activity.Name
	}
	return &model.RechargeBatch{
		UserID:                  order.UserID,
		RechargeOrderID:         &orderID,
		ActivityID:              &activityID,
		ActivityNameSnapshot:    activityName,
		ActivityVersionSnapshot: normalizeRechargeActivityRuleVersion(order.ActivityVersionSnapshot),
		BatchType:               model.RechargeBatchTypeRecharge,
		PrincipalTotal:          order.RechargeAmount,
		GiftTotal:               order.GiftAmount,
		PrincipalRemaining:      principalRemaining,
		GiftRemaining:           giftRemaining,
		PrincipalConsumeWeight:  normalizeConsumeWeight(order.PrincipalConsumeWeight, order.RechargeAmount),
		GiftConsumeWeight:       normalizeConsumeWeight(order.GiftConsumeWeight, order.GiftAmount),
	}
}

func isLegacyRechargeOrderPlanError(err error) bool {
	return err != nil && (errors.Is(err, errLegacyRechargeOrderAmbiguous) || errors.Is(err, errLegacyRechargeOrderNoRemaining) || errors.Is(err, errLegacyRechargeOrderBalanceMismatch))
}

func applyAdminRechargeCleanupState(resp *response.AdminRechargeOrderResp, order *model.RechargeOrder, refund model.RefundRequest, batch *model.RechargeBatch, fallbackReason string) {
	if resp == nil {
		return
	}
	if batch == nil {
		if trimmedReason := strings.TrimSpace(fallbackReason); trimmedReason != "" {
			resp.CleanupUnavailableReason = trimmedReason
		} else {
			resp.CleanupUnavailableReason = "充值批次不存在"
		}
		resp.CanCleanupRemaining = false
		return
	}
	if order == nil || order.Status != model.RechargeOrderStatusPaid {
		resp.CleanupUnavailableReason = "当前充值订单状态不可清理"
		resp.CanCleanupRemaining = false
		return
	}
	if resp.CurrentRemainingAmount <= 0 {
		resp.CleanupUnavailableReason = "当前充值订单无可清理剩余余额"
		resp.CanCleanupRemaining = false
		return
	}
	if isBlockingRechargeRefundStatus(refund.Status) {
		resp.CleanupUnavailableReason = "当前充值订单已有待处理或已完成退款，不能测试清理"
		resp.CanCleanupRemaining = false
		return
	}
	resp.CleanupUnavailableReason = ""
	resp.CanCleanupRemaining = true
}

func applyAdminRechargeOrderListState(resp *response.AdminRechargeOrderResp, order *model.RechargeOrder, refund model.RefundRequest, batch *model.RechargeBatch) {
	if resp == nil || order == nil {
		return
	}
	if order.Status == model.RechargeOrderStatusPending {
		resp.BatchStatus = "pending"
		resp.BatchStatusText = "待支付未生成批次"
		resp.SuggestedAction = "wait_pay"
		resp.SuggestedActionText = "等待支付或取消，无需运营介入"
		return
	}
	if order.Status == model.RechargeOrderStatusCancelled {
		resp.BatchStatus = "cancelled"
		resp.BatchStatusText = "已取消"
		resp.SuggestedAction = "no_action"
		resp.SuggestedActionText = "订单已取消，无需处理"
		return
	}
	if batch == nil {
		resp.BatchStatus = "legacy_missing"
		resp.BatchStatusText = "无批次旧单"
		resp.SuggestedAction = "handle_historical_balance"
		resp.SuggestedActionText = "去用户详情处理历史余额"
		return
	}
	if resp.CurrentRemainingAmount <= 0 {
		resp.BatchStatus = "depleted"
		resp.BatchStatusText = "批次已耗尽"
		resp.SuggestedAction = "no_action"
		resp.SuggestedActionText = "当前已无剩余，无需处理"
		return
	}
	resp.BatchStatus = "normal"
	resp.BatchStatusText = "正常批次"
	if isBlockingRechargeRefundStatus(refund.Status) {
		resp.SuggestedAction = "handle_refund"
		resp.SuggestedActionText = "先处理退款申请"
		return
	}
	if batch.PrincipalRemaining > 0 {
		resp.SuggestedAction = "refund_or_cleanup"
		resp.SuggestedActionText = "可申请退款；测试场景可作废此批次剩余"
		return
	}
	resp.SuggestedAction = "cleanup_only"
	resp.SuggestedActionText = "如仅需清空赠送，作废此批次剩余"
}

func applyAdminRechargeRefundState(resp *response.AdminRechargeOrderResp, order *model.RechargeOrder, refund model.RefundRequest, batch *model.RechargeBatch) {
	if resp == nil {
		return
	}
	if order == nil || order.Status != model.RechargeOrderStatusPaid {
		resp.RefundUnavailableReason = "当前充值订单状态不可退款"
		resp.CanCreateRefund = false
		return
	}
	if isBlockingRechargeRefundStatus(refund.Status) {
		resp.RefundUnavailableReason = "当前充值订单已有待处理或已完成退款，不能再次申请"
		resp.CanCreateRefund = false
		return
	}
	if batch == nil {
		resp.RefundUnavailableReason = "当前充值订单没有充值批次，不能直接申请退款"
		resp.CanCreateRefund = false
		return
	}
	if batch.PrincipalRemaining <= 0 {
		resp.RefundUnavailableReason = "当前充值订单无可退本金"
		resp.CanCreateRefund = false
		return
	}
	resp.RefundUnavailableReason = ""
	resp.CanCreateRefund = true
}

func buildAdminRechargeCleanupRemark(adminID uint, orderNo, remark string) string {
	base := "后台作废此批次剩余"
	if trimmedRemark := strings.TrimSpace(remark); trimmedRemark != "" {
		base += "：" + trimmedRemark
	}
	if trimmedOrderNo := strings.TrimSpace(orderNo); trimmedOrderNo != "" {
		base += "（充值单 " + trimmedOrderNo + "，管理员#" + strconv.FormatUint(uint64(adminID), 10) + "）"
	}
	return base
}

func isBlockingRechargeRefundStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case model.RefundStatusPendingReview, model.RefundStatusApproved, model.RefundStatusProcessing, model.RefundStatusSuccess:
		return true
	default:
		return false
	}
}

func validateAdminRechargeOrderStatus(status *int) error {
	if status == nil {
		return nil
	}
	if *status != model.RechargeOrderStatusPending && *status != model.RechargeOrderStatusPaid && *status != model.RechargeOrderStatusCancelled {
		return errors.New("充值订单状态无效")
	}
	return nil
}

func validateAdminRechargeOrderSuggestedAction(suggestedAction string) error {
	switch strings.TrimSpace(suggestedAction) {
	case "", "wait_pay", "no_action", "handle_historical_balance", "handle_refund", "refund_or_cleanup", "cleanup_only":
		return nil
	default:
		return errors.New("充值订单建议动作无效")
	}
}

func parseAdminRechargeOrderTimeRange(startAtRaw, endAtRaw string) (*time.Time, *time.Time, error) {
	startAt, err := parseAdminDateTimeValue(startAtRaw)
	if err != nil {
		return nil, nil, errors.New("充值订单时间范围无效")
	}
	endAt, err := parseAdminDateTimeValue(endAtRaw)
	if err != nil {
		return nil, nil, errors.New("充值订单时间范围无效")
	}
	if startAt != nil && endAt != nil && startAt.After(*endAt) {
		return nil, nil, errors.New("充值订单时间范围无效")
	}
	return startAt, endAt, nil
}

func parseAdminDateTimeValue(raw string) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	value, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, time.Local)
	if err == nil {
		return &value, nil
	}
	value, err = time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func adminRechargeOrderStatusText(status int) string {
	switch status {
	case model.RechargeOrderStatusPending:
		return "待支付"
	case model.RechargeOrderStatusPaid:
		return "已支付"
	case model.RechargeOrderStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

func adminRechargeOrderPayChannelText(channel string) string {
	switch channel {
	case "wechat":
		return "微信支付"
	default:
		return channel
	}
}

func adminRefundStatusText(status string) string {
	switch status {
	case model.RefundStatusPendingReview:
		return "待审核"
	case model.RefundStatusApproved:
		return "已通过"
	case model.RefundStatusProcessing:
		return "处理中"
	case model.RefundStatusSuccess:
		return "已退款"
	case model.RefundStatusRejected:
		return "已驳回"
	case model.RefundStatusFailed:
		return "退款失败"
	case model.RefundStatusCancelled:
		return "已撤销"
	case model.RefundStatusException:
		return "异常"
	default:
		return "-"
	}
}

func formatAdminTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
