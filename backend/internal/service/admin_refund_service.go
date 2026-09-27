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
	"deskorder/pkg/database"
	"deskorder/pkg/wechat"

	"gorm.io/gorm"
)

func AdminGetRechargeRefundRequests(page, size int, status, requestChannel string) ([]response.AdminRefundRequestResp, int64, error) {
	list, total, err := repository.ListRefundRequests(page, size, model.RefundReviewScopeRecharge, strings.TrimSpace(status), strings.TrimSpace(requestChannel))
	if err != nil {
		return nil, 0, errors.New("获取充值退款申请列表失败")
	}
	respList := make([]response.AdminRefundRequestResp, 0, len(list))
	for i := range list {
		item, buildErr := buildAdminRefundRequestResp(&list[i])
		if buildErr != nil {
			return nil, 0, buildErr
		}
		respList = append(respList, item)
	}
	return respList, total, nil
}

func AdminGetRechargeRefundRequestDetail(requestID uint) (*response.AdminRefundRequestResp, error) {
	item, err := repository.GetRefundRequestByID(requestID)
	if err != nil {
		return nil, errors.New("退款申请不存在")
	}
	if item.ReviewScope != model.RefundReviewScopeRecharge {
		return nil, errors.New("退款申请不存在")
	}
	resp, err := buildAdminRefundRequestResp(item)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func AdminCreateRechargeRefundRequest(adminID uint, req *request.AdminCreateRechargeRefundReq) (*response.AdminRefundRequestResp, error) {
	if req == nil {
		return nil, errors.New("参数错误")
	}
	rechargeOrderID, err := resolveRechargeRefundOrderID(req)
	if err != nil {
		return nil, err
	}
	ctx, err := loadRechargeRefundContext(rechargeOrderID)
	if err != nil {
		return nil, err
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
	var item *model.RefundRequest
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		requestNote := firstNonEmptyString(strings.TrimSpace(req.RequestNote), strings.TrimSpace(req.ReviewNote))
		item, err = createRechargeRefundRequestWithDB(tx, ctx, model.RefundRequestChannelAdmin, requestNote)
		return err
	}); err != nil {
		return nil, errors.New("创建充值退款申请失败")
	}
	resp, err := buildAdminRefundRequestResp(item)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func AdminReviewRechargeRefundRequest(adminID, requestID uint, req *request.AdminReviewRechargeRefundReq) (*response.AdminRefundRequestResp, error) {
	if req == nil {
		return nil, errors.New("参数错误")
	}
	action := strings.TrimSpace(strings.ToLower(req.Action))
	switch action {
	case "reject":
		return adminRejectRechargeRefundRequest(adminID, requestID, req)
	case "complete":
		return adminCompleteRechargeRefundRequest(adminID, requestID, req)
	default:
		return nil, errors.New("退款操作无效")
	}
}

func adminRejectRechargeRefundRequest(adminID, requestID uint, req *request.AdminReviewRechargeRefundReq) (*response.AdminRefundRequestResp, error) {
	if strings.TrimSpace(req.RejectReason) == "" {
		return nil, errors.New("请输入驳回原因")
	}
	var item *model.RefundRequest
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		item, err = repository.GetRefundRequestByIDWithDB(tx, requestID)
		if err != nil {
			return errors.New("退款申请不存在")
		}
		if item.ReviewScope != model.RefundReviewScopeRecharge {
			return errors.New("退款申请不存在")
		}
		if item.Status != model.RefundStatusPendingReview {
			return errors.New("当前退款申请状态不可驳回")
		}
		now := time.Now()
		item.Status = model.RefundStatusRejected
		item.RejectReason = strings.TrimSpace(req.RejectReason)
		item.ReviewNote = strings.TrimSpace(req.ReviewNote)
		item.ReviewerRole = "admin"
		item.ReviewerID = adminID
		item.ReviewedAt = &now
		if err := repository.UpdateRefundRequestWithDB(tx, item); err != nil {
			return err
		}
		return createUserRefundNotificationWithDB(tx, item.UserID, model.RefundSourceTypeRecharge, item.SourceID, "充值退款未通过审核", buildUserRechargeRefundResultText(item))
	}); err != nil {
		return nil, errors.New("驳回充值退款申请失败")
	}
	triggerRefundReviewSubscribeMessage(item)
	resp, err := buildAdminRefundRequestResp(item)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func adminCompleteRechargeRefundRequest(adminID, requestID uint, req *request.AdminReviewRechargeRefundReq) (*response.AdminRefundRequestResp, error) {
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		item, err := repository.GetRefundRequestByIDWithDB(tx, requestID)
		if err != nil {
			return errors.New("退款申请不存在")
		}
		if item.ReviewScope != model.RefundReviewScopeRecharge {
			return errors.New("退款申请不存在")
		}
		if item.Status != model.RefundStatusPendingReview {
			return errors.New("当前退款申请状态不可完成")
		}

		order, err := repository.GetRechargeOrderByID(item.SourceID)
		if err != nil {
			return errors.New("充值订单不存在")
		}
		if order.Status != model.RechargeOrderStatusPaid {
			return errors.New("当前充值订单状态不可退款")
		}
		batch, err := repository.GetRechargeBatchByRechargeOrderIDWithDB(tx, order.ID)
		if err != nil {
			return errors.New("充值批次不存在")
		}
		principalAmount := batch.PrincipalRemaining
		giftAmount := batch.GiftRemaining
		if principalAmount <= 0 {
			return errors.New("当前充值订单无可退本金")
		}
		totalAmount := principalAmount + giftAmount
		completion, err := resolveRechargeRefundCompletion(order, principalAmount, req)
		if err != nil {
			return err
		}

		balance, err := repository.GetUserBalanceByUserIDWithDB(tx, order.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户余额不存在")
		}
		if err != nil {
			return err
		}
		if balance.BalanceAmount < totalAmount {
			return errors.New("用户当前余额不足，无法完成充值退款，请人工处理")
		}

		now := time.Now()
		finishedAt := (*time.Time)(nil)
		if completion.ExecutionStatus == model.RefundExecutionStatusSuccess {
			finishedAt = &now
		}
		execution := &model.RefundExecution{
			RefundRequestID:        item.ID,
			ExecutionNo:            fmt.Sprintf("RE%d%d", now.UnixNano(), item.ID),
			Channel:                completion.Channel,
			OriginalPaySourceType:  model.RefundSourceTypeRecharge,
			OriginalPaySourceID:    order.ID,
			RechargeBatchID:        &batch.ID,
			ExecuteTotalAmount:     totalAmount,
			ExecutePrincipalAmount: principalAmount,
			ExecuteGiftVoidAmount:  giftAmount,
			Status:                 completion.ExecutionStatus,
			ExternalRefundNo:       completion.ExternalRefundNo,
			ExternalTransactionID:  completion.ExternalTransactionID,
			ExternalResponseCode:   completion.ExternalResponseCode,
			ExternalResponseMsg:    completion.ExternalResponseMsg,
			StartedAt:              &now,
			FinishedAt:             finishedAt,
		}
		if err := repository.CreateRefundExecutionWithDB(tx, execution); err != nil {
			return err
		}

		if err := applyRechargeRefundSettlementWithDB(tx, order, batch, balance, principalAmount, giftAmount, completion.BalanceRemark); err != nil {
			return err
		}

		item.Status = completion.RequestStatus
		item.ApprovedTotalAmount = totalAmount
		item.ApprovedPrincipalAmount = principalAmount
		item.ApprovedGiftVoidAmount = giftAmount
		item.ReviewNote = strings.TrimSpace(req.ReviewNote)
		item.ReviewerRole = "admin"
		item.ReviewerID = adminID
		item.ReviewedAt = &now
		if err := repository.UpdateRefundRequestWithDB(tx, item); err != nil {
			return err
		}
		title := "充值退款处理中"
		if item.Status == model.RefundStatusSuccess {
			title = "充值退款已完成"
		}
		return createUserRefundNotificationWithDB(tx, item.UserID, model.RefundSourceTypeRecharge, item.SourceID, title, buildUserRechargeRefundResultText(item))
	}); err != nil {
		return nil, err
	}
	updatedItem, getErr := repository.GetRefundRequestByID(requestID)
	if getErr == nil {
		triggerRefundReviewSubscribeMessage(updatedItem)
		triggerRefundResultSubscribeMessage(updatedItem)
	}

	resp, err := AdminGetRechargeRefundRequestDetail(requestID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func AdminSyncRechargeRefundRequestStatus(requestID uint) (*response.AdminRefundRequestResp, error) {
	item, err := repository.GetRefundRequestByID(requestID)
	if err != nil {
		return nil, errors.New("退款申请不存在")
	}
	if item.ReviewScope != model.RefundReviewScopeRecharge {
		return nil, errors.New("退款申请不存在")
	}
	if item.Status != model.RefundStatusProcessing {
		return nil, errors.New("当前退款申请状态不可同步")
	}
	executions, err := repository.GetRefundExecutionsByRequestID(item.ID)
	if err != nil || len(executions) == 0 {
		return nil, errors.New("退款执行记录不存在")
	}
	latestExecution := executions[len(executions)-1]
	if latestExecution.Status != model.RefundExecutionStatusProcessing {
		return nil, errors.New("当前退款执行状态不可同步")
	}
	queryResult, err := queryRechargeRefundStatus(item, &latestExecution)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		currentItem, err := repository.GetRefundRequestByIDWithDB(tx, requestID)
		if err != nil {
			return errors.New("退款申请不存在")
		}
		if currentItem.Status != model.RefundStatusProcessing {
			return errors.New("当前退款申请状态不可同步")
		}
		currentExecution, err := repository.GetLatestRefundExecutionByRequestIDWithDB(tx, currentItem.ID)
		if err != nil {
			return errors.New("退款执行记录不存在")
		}
		if currentExecution.Status != model.RefundExecutionStatusProcessing {
			return errors.New("当前退款执行状态不可同步")
		}
		currentExecution.RetryCount++
		currentExecution.ExternalResponseCode = queryResult.ExternalResponseCode
		currentExecution.ExternalResponseMsg = queryResult.ExternalResponseMsg
		if queryResult.ExternalTransactionID != "" {
			currentExecution.ExternalTransactionID = queryResult.ExternalTransactionID
		}
		now := time.Now()

		switch queryResult.MappedStatus {
		case model.RefundExecutionStatusProcessing:
			return repository.UpdateRefundExecutionWithDB(tx, currentExecution)
		case model.RefundExecutionStatusSuccess:
			currentExecution.Status = model.RefundExecutionStatusSuccess
			currentExecution.FinishedAt = &now
			if err := repository.UpdateRefundExecutionWithDB(tx, currentExecution); err != nil {
				return err
			}
			currentItem.Status = model.RefundStatusSuccess
			if err := repository.UpdateRefundRequestWithDB(tx, currentItem); err != nil {
				return err
			}
			return createUserRefundNotificationWithDB(tx, currentItem.UserID, model.RefundSourceTypeRecharge, currentItem.SourceID, "充值退款已完成", buildUserRechargeRefundResultText(currentItem))
		case model.RefundExecutionStatusFailed:
			currentExecution.Status = model.RefundExecutionStatusFailed
			currentExecution.FinishedAt = &now
			if err := repository.UpdateRefundExecutionWithDB(tx, currentExecution); err != nil {
				return err
			}
			order, err := repository.GetRechargeOrderByID(currentItem.SourceID)
			if err != nil {
				return errors.New("充值订单不存在")
			}
			batch, err := repository.GetRechargeBatchByRechargeOrderIDWithDB(tx, order.ID)
			if err != nil {
				return errors.New("充值批次不存在")
			}
			balance, err := repository.GetUserBalanceByUserIDWithDB(tx, order.UserID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户余额不存在")
			}
			if err != nil {
				return err
			}
			if err := rollbackRechargeRefundSettlementWithDB(tx, order, batch, balance, currentExecution.ExecutePrincipalAmount, currentExecution.ExecuteGiftVoidAmount, "微信充值退款失败，回滚预扣余额"); err != nil {
				return err
			}
			currentItem.Status = model.RefundStatusFailed
			if err := repository.UpdateRefundRequestWithDB(tx, currentItem); err != nil {
				return err
			}
			return createUserRefundNotificationWithDB(tx, currentItem.UserID, model.RefundSourceTypeRecharge, currentItem.SourceID, "充值退款处理失败", buildUserRechargeRefundResultText(currentItem))
		default:
			return errors.New("微信退款状态未知，请稍后重试")
		}
	}); err != nil {
		return nil, err
	}
	updatedItem, getErr := repository.GetRefundRequestByID(requestID)
	if getErr == nil {
		triggerRefundResultSubscribeMessage(updatedItem)
	}

	return AdminGetRechargeRefundRequestDetail(requestID)
}

func buildAdminRefundRequestResp(item *model.RefundRequest) (response.AdminRefundRequestResp, error) {
	resp := response.AdminRefundRequestResp{
		ID:                       item.ID,
		RequestNo:                item.RequestNo,
		SourceType:               item.SourceType,
		ReviewScope:              item.ReviewScope,
		RequestChannel:           item.RequestChannel,
		RequestChannelText:       adminRefundRequestChannelText(item.RequestChannel),
		SourceID:                 item.SourceID,
		UserID:                   item.UserID,
		Status:                   item.Status,
		RequestedTotalAmount:     item.RequestedTotalAmount,
		RequestedPrincipalAmount: item.RequestedPrincipalAmount,
		RequestedGiftVoidAmount:  item.RequestedGiftVoidAmount,
		ApprovedTotalAmount:      item.ApprovedTotalAmount,
		ApprovedPrincipalAmount:  item.ApprovedPrincipalAmount,
		ApprovedGiftVoidAmount:   item.ApprovedGiftVoidAmount,
		RequestNote:              item.RequestNote,
		ReviewNote:               item.ReviewNote,
		RejectReason:             item.RejectReason,
		CreatedAt:                item.CreatedAt.Format("2006-01-02 15:04:05"),
		ReviewedAt:               formatNullableAdminTime(item.ReviewedAt),
	}
	user, err := repository.GetUserByID(item.UserID)
	if err != nil {
		return response.AdminRefundRequestResp{}, errors.New("查询退款申请用户失败")
	}
	resp.UserNickname = user.Nickname
	resp.UserPhone = user.Phone

	if item.SourceType == model.RefundSourceTypeRecharge {
		order, err := repository.GetRechargeOrderByID(item.SourceID)
		if err != nil {
			return response.AdminRefundRequestResp{}, errors.New("查询退款申请充值订单失败")
		}
		resp.RechargeOrderID = order.ID
		resp.RechargeOrderNo = order.OrderNo
		batch, batchErr := repository.GetRechargeBatchByRechargeOrderID(order.ID)
		if batchErr == nil && batch != nil {
			resp.CurrentPrincipalRemaining = batch.PrincipalRemaining
			resp.CurrentGiftRemaining = batch.GiftRemaining
		}
	}
	balance, err := repository.GetUserBalanceByUserID(item.UserID)
	if err == nil && balance != nil {
		resp.CurrentBalanceAmount = balance.BalanceAmount
	}
	last, err := repository.GetLatestRefundExecutionByRequestID(item.ID)
	if err == nil && last != nil {
		resp.ExecutionStatus = last.Status
		resp.SyncRetryCount = last.RetryCount
		resp.LastSyncAt = formatAdminExecutionSyncTime(last)
		resp.NextAutoSyncAt = formatAdminTime(calculateNextRechargeRefundAutoSyncAt(last))
		resp.AutoSyncPaused = isRechargeRefundAutoSyncPaused(last)
		resp.ExternalRefundNo = last.ExternalRefundNo
		resp.ExternalResponseCode = last.ExternalResponseCode
		resp.ExternalResponseMsg = last.ExternalResponseMsg
	}
	return resp, nil
}

type rechargeRefundContext struct {
	order   *model.RechargeOrder
	batch   *model.RechargeBatch
	balance *model.UserBalance
}

func createRechargeRefundRequestWithDB(tx *gorm.DB, ctx *rechargeRefundContext, requestChannel, requestNote string) (*model.RefundRequest, error) {
	if tx == nil || ctx == nil || ctx.order == nil || ctx.batch == nil {
		return nil, errors.New("充值退款上下文无效")
	}
	now := time.Now()
	item := &model.RefundRequest{
		RequestNo:                fmt.Sprintf("RR%d%d", now.UnixNano(), ctx.order.ID),
		SourceType:               model.RefundSourceTypeRecharge,
		SourceID:                 ctx.order.ID,
		UserID:                   ctx.order.UserID,
		RequestChannel:           requestChannel,
		ReviewScope:              model.RefundReviewScopeRecharge,
		Status:                   model.RefundStatusPendingReview,
		RequestedTotalAmount:     ctx.batch.PrincipalRemaining + ctx.batch.GiftRemaining,
		RequestedPrincipalAmount: ctx.batch.PrincipalRemaining,
		RequestedGiftVoidAmount:  ctx.batch.GiftRemaining,
		RequestNote:              strings.TrimSpace(requestNote),
		ReviewerRole:             "",
		ReviewerID:               0,
		ReviewedAt:               nil,
	}
	if err := repository.CreateRefundRequestWithDB(tx, item); err != nil {
		return nil, err
	}
	if err := createUserRefundNotificationWithDB(tx, item.UserID, model.RefundSourceTypeRecharge, item.SourceID, "充值退款已受理", "平台已收到你的充值退款申请，当前状态为待审核。"); err != nil {
		return nil, err
	}
	return item, nil
}

func adminRefundRequestChannelText(channel string) string {
	switch channel {
	case model.RefundRequestChannelUser:
		return "用户发起"
	case model.RefundRequestChannelAdmin:
		return "后台发起"
	case model.RefundRequestChannelMerchant:
		return "商家发起"
	default:
		return "未知来源"
	}
}

func loadRechargeRefundContext(rechargeOrderID uint) (*rechargeRefundContext, error) {
	order, err := repository.GetRechargeOrderByID(rechargeOrderID)
	if err != nil {
		return nil, errors.New("充值订单不存在")
	}
	if order.Status != model.RechargeOrderStatusPaid {
		return nil, errors.New("当前充值订单状态不可退款")
	}
	batch, err := repository.GetRechargeBatchByRechargeOrderID(order.ID)
	if err != nil {
		return nil, errors.New("充值批次不存在")
	}
	balance, err := repository.GetOrCreateUserBalanceByUserID(order.UserID)
	if err != nil {
		return nil, errors.New("查询用户余额失败")
	}
	return &rechargeRefundContext{order: order, batch: batch, balance: balance}, nil
}

func resolveRechargeRefundOrderID(req *request.AdminCreateRechargeRefundReq) (uint, error) {
	if req == nil {
		return 0, errors.New("参数错误")
	}
	if req.RechargeOrderID > 0 {
		return req.RechargeOrderID, nil
	}
	rechargeOrderNo := strings.TrimSpace(req.RechargeOrderNo)
	if rechargeOrderNo == "" {
		return 0, errors.New("请输入充值订单ID或充值单号")
	}
	order, err := repository.GetRechargeOrderByOrderNo(rechargeOrderNo)
	if err != nil {
		return 0, errors.New("充值订单不存在")
	}
	return order.ID, nil
}

type rechargeRefundCompletion struct {
	Channel               string
	ExecutionStatus       string
	RequestStatus         string
	ExternalRefundNo      string
	ExternalTransactionID string
	ExternalResponseCode  string
	ExternalResponseMsg   string
	BalanceRemark         string
}

func resolveRechargeRefundCompletion(order *model.RechargeOrder, principalAmount int, req *request.AdminReviewRechargeRefundReq) (*rechargeRefundCompletion, error) {
	manualRefundNo := strings.TrimSpace(req.ExternalRefundNo)
	if manualRefundNo != "" {
		return &rechargeRefundCompletion{
			Channel:             model.RefundExecutionChannelWechat,
			ExecutionStatus:     model.RefundExecutionStatusSuccess,
			RequestStatus:       model.RefundStatusSuccess,
			ExternalRefundNo:    manualRefundNo,
			ExternalResponseMsg: strings.TrimSpace(req.ExternalResponseMsg),
			BalanceRemark:       "平台手工充值退款扣减余额",
		}, nil
	}

	if normalizePayMode(config.AppConfig.WeChat.PayMode) == payModeMock {
		return &rechargeRefundCompletion{
			Channel:              model.RefundExecutionChannelWechat,
			ExecutionStatus:      model.RefundExecutionStatusProcessing,
			RequestStatus:        model.RefundStatusProcessing,
			ExternalRefundNo:     fmt.Sprintf("MOCKRF%d%d", time.Now().UnixNano(), order.ID),
			ExternalResponseCode: "MOCK_ACCEPTED",
			ExternalResponseMsg:  "mock微信退款已受理，待确认",
			BalanceRemark:        "mock微信充值退款处理中，预扣余额",
		}, nil
	}

	if order.PayChannel != model.PaymentChannelWechat {
		return nil, errors.New("当前充值订单不是微信支付，请填写外部退款单号后手工完成")
	}
	if config.AppConfig == nil {
		return nil, errors.New("微信退款未配置完成，请先配置退款证书或填写外部退款单号手工完成")
	}
	cfg := config.AppConfig.WeChat
	if cfg.UserAppID == "" || cfg.MchID == "" || cfg.MchAPIKey == "" || strings.TrimSpace(cfg.RefundCertP12Path) == "" {
		return nil, errors.New("微信退款未配置完成，请先配置退款证书或填写外部退款单号手工完成")
	}
	refundResp, err := wechat.Refund(&wechat.RefundReq{
		AppID:       cfg.UserAppID,
		MchID:       cfg.MchID,
		OutTradeNo:  order.OrderNo,
		OutRefundNo: fmt.Sprintf("RF%d%d", time.Now().UnixNano(), order.ID),
		TotalFee:    order.PayAmount,
		RefundFee:   principalAmount,
		OpUserID:    cfg.MchID,
	}, cfg.MchAPIKey, cfg.RefundCertP12Path)
	if err != nil {
		return nil, fmt.Errorf("微信退款失败: %w", err)
	}
	externalRefundNo := strings.TrimSpace(refundResp.OutRefundNo)
	if externalRefundNo == "" {
		externalRefundNo = fmt.Sprintf("RF%d%d", time.Now().UnixNano(), order.ID)
	}
	return &rechargeRefundCompletion{
		Channel:               model.RefundExecutionChannelWechat,
		ExecutionStatus:       model.RefundExecutionStatusProcessing,
		RequestStatus:         model.RefundStatusProcessing,
		ExternalRefundNo:      externalRefundNo,
		ExternalTransactionID: strings.TrimSpace(refundResp.RefundID),
		ExternalResponseCode:  strings.TrimSpace(refundResp.ResultCode),
		ExternalResponseMsg:   firstNonEmptyString(strings.TrimSpace(refundResp.ErrCodeDes), strings.TrimSpace(refundResp.ReturnMsg), "微信退款已受理"),
		BalanceRemark:         "微信充值退款处理中，预扣余额",
	}, nil
}

type rechargeRefundQueryResult struct {
	MappedStatus          string
	ExternalTransactionID string
	ExternalResponseCode  string
	ExternalResponseMsg   string
}

func queryRechargeRefundStatus(item *model.RefundRequest, execution *model.RefundExecution) (*rechargeRefundQueryResult, error) {
	if item == nil || execution == nil {
		return nil, errors.New("退款执行记录不存在")
	}
	if normalizePayMode(config.AppConfig.WeChat.PayMode) == payModeMock {
		return &rechargeRefundQueryResult{
			MappedStatus:         model.RefundExecutionStatusSuccess,
			ExternalResponseCode: "MOCK_SUCCESS",
			ExternalResponseMsg:  "mock微信退款成功",
		}, nil
	}
	if config.AppConfig == nil {
		return nil, errors.New("微信退款查询未配置完成，请先配置后端支付参数")
	}
	cfg := config.AppConfig.WeChat
	if cfg.UserAppID == "" || cfg.MchID == "" || cfg.MchAPIKey == "" {
		return nil, errors.New("微信退款查询未配置完成，请先配置后端支付参数")
	}
	queryResp, err := wechat.RefundQuery(&wechat.RefundQueryReq{
		AppID:       cfg.UserAppID,
		MchID:       cfg.MchID,
		OutRefundNo: execution.ExternalRefundNo,
	}, cfg.MchAPIKey)
	if err != nil {
		return nil, fmt.Errorf("微信退款状态查询失败: %w", err)
	}
	result := &rechargeRefundQueryResult{
		ExternalTransactionID: firstNonEmptyString(strings.TrimSpace(queryResp.RefundID), strings.TrimSpace(queryResp.TransactionID)),
		ExternalResponseCode:  strings.TrimSpace(queryResp.RefundStatus),
		ExternalResponseMsg:   firstNonEmptyString(strings.TrimSpace(queryResp.RefundSuccessTime), strings.TrimSpace(queryResp.ErrCodeDes), strings.TrimSpace(queryResp.ReturnMsg)),
	}
	switch strings.ToUpper(strings.TrimSpace(queryResp.RefundStatus)) {
	case "SUCCESS":
		result.MappedStatus = model.RefundExecutionStatusSuccess
	case "PROCESSING":
		result.MappedStatus = model.RefundExecutionStatusProcessing
	case "CHANGE", "REFUNDCLOSE":
		result.MappedStatus = model.RefundExecutionStatusFailed
	default:
		result.MappedStatus = model.RefundExecutionStatusFailed
		if result.ExternalResponseMsg == "" {
			result.ExternalResponseMsg = "微信返回未知退款状态"
		}
	}
	return result, nil
}

func applyRechargeRefundSettlementWithDB(tx *gorm.DB, order *model.RechargeOrder, batch *model.RechargeBatch, balance *model.UserBalance, principalAmount, giftAmount int, remark string) error {
	totalAmount := principalAmount + giftAmount
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
	if batch.PrincipalRemaining >= principalAmount {
		batch.PrincipalRemaining -= principalAmount
	} else {
		batch.PrincipalRemaining = 0
	}
	if batch.GiftRemaining >= giftAmount {
		batch.GiftRemaining -= giftAmount
	} else {
		batch.GiftRemaining = 0
	}
	if err := repository.SaveRechargeBatchWithDB(tx, batch); err != nil {
		return err
	}
	return repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
		UserID:                order.UserID,
		ChangeAmount:          -totalAmount,
		PrincipalChangeAmount: -principalAmount,
		GiftChangeAmount:      -giftAmount,
		BalanceAfter:          balance.BalanceAmount,
		BizType:               model.BalanceTransactionTypeRefund,
		SourceType:            model.BalanceTransactionSourceRechargeOrder,
		SourceID:              order.ID,
		RelatedBatchID:        &batch.ID,
		Remark:                remark,
	})
}

func rollbackRechargeRefundSettlementWithDB(tx *gorm.DB, order *model.RechargeOrder, batch *model.RechargeBatch, balance *model.UserBalance, principalAmount, giftAmount int, remark string) error {
	totalAmount := principalAmount + giftAmount
	balance.BalanceAmount += totalAmount
	balance.TotalRechargeAmount += principalAmount
	balance.TotalGiftAmount += giftAmount
	if err := repository.SaveUserBalanceWithDB(tx, balance); err != nil {
		return err
	}
	batch.PrincipalRemaining += principalAmount
	if batch.PrincipalRemaining > batch.PrincipalTotal {
		batch.PrincipalRemaining = batch.PrincipalTotal
	}
	batch.GiftRemaining += giftAmount
	if batch.GiftRemaining > batch.GiftTotal {
		batch.GiftRemaining = batch.GiftTotal
	}
	if err := repository.SaveRechargeBatchWithDB(tx, batch); err != nil {
		return err
	}
	return repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
		UserID:                order.UserID,
		ChangeAmount:          totalAmount,
		PrincipalChangeAmount: principalAmount,
		GiftChangeAmount:      giftAmount,
		BalanceAfter:          balance.BalanceAmount,
		BizType:               model.BalanceTransactionTypeAdjust,
		SourceType:            model.BalanceTransactionSourceRechargeOrder,
		SourceID:              order.ID,
		RelatedBatchID:        &batch.ID,
		Remark:                remark,
	})
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func formatAdminExecutionSyncTime(item *model.RefundExecution) string {
	if item == nil {
		return ""
	}
	if item.RetryCount <= 0 && item.Status == model.RefundExecutionStatusProcessing {
		return ""
	}
	return item.UpdatedAt.Format("2006-01-02 15:04:05")
}

func formatNullableAdminTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
