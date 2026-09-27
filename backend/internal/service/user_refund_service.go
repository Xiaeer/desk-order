package service

import (
	"errors"
	"fmt"

	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func GetUserRefundRecords(userID uint, page, size int) ([]response.UserRefundRecordResp, int64, error) {
	list, total, err := repository.ListRefundRequestsByUserID(userID, page, size)
	if err != nil {
		return nil, 0, errors.New("查询退款记录失败")
	}
	respList := make([]response.UserRefundRecordResp, 0, len(list))
	for i := range list {
		respList = append(respList, buildUserRefundRecordResp(&list[i]))
	}
	return respList, total, nil
}

func GetUserRefundNotifications(userID uint, page, size int) ([]response.UserNotificationResp, int64, error) {
	list, total, err := repository.ListUserNotificationsByCategory(userID, model.UserNotificationCategoryRefund, page, size)
	if err != nil {
		return nil, 0, errors.New("查询退款通知失败")
	}
	respList := make([]response.UserNotificationResp, 0, len(list))
	for i := range list {
		respList = append(respList, response.UserNotificationResp{
			ID:          list[i].ID,
			Category:    list[i].Category,
			Title:       list[i].Title,
			Content:     list[i].Content,
			RelatedType: list[i].RelatedType,
			RelatedID:   list[i].RelatedID,
			IsRead:      list[i].IsRead,
			CreatedAt:   list[i].CreatedAt,
		})
	}
	return respList, total, nil
}

func ReadAllUserRefundNotifications(userID uint) error {
	if err := repository.MarkAllUserNotificationsReadByCategory(userID, model.UserNotificationCategoryRefund); err != nil {
		return errors.New("更新退款通知已读失败")
	}
	return nil
}

func createUserRefundNotificationWithDB(db *gorm.DB, userID uint, relatedType string, relatedID uint, title, content string) error {
	if db == nil || userID == 0 {
		return nil
	}
	item := &model.UserNotification{
		UserID:      userID,
		Category:    model.UserNotificationCategoryRefund,
		Title:       title,
		Content:     content,
		RelatedType: relatedType,
		RelatedID:   relatedID,
	}
	return repository.CreateUserNotificationWithDB(db, item)
}

func buildUserRefundRecordResp(item *model.RefundRequest) response.UserRefundRecordResp {
	resp := response.UserRefundRecordResp{}
	if item == nil {
		return resp
	}
	resp.ID = item.ID
	resp.RequestNo = item.RequestNo
	resp.SourceType = item.SourceType
	resp.SourceID = item.SourceID
	resp.RequestNote = item.RequestNote
	resp.CanCancel = canUserCancelRefundRequest(item)
	if resp.CanCancel {
		resp.CancelActionText = "撤回申请"
	}
	resp.Status = item.Status
	resp.StatusText = userRefundStatusText(item.Status)
	resp.RequestedTotalAmount = item.RequestedTotalAmount
	resp.ApprovedTotalAmount = item.ApprovedTotalAmount
	resp.ResultText = buildUserRefundResultText(item)
	resp.CreatedAt = item.CreatedAt
	resp.UpdatedAt = item.UpdatedAt

	switch item.SourceType {
	case model.RefundSourceTypeRecharge:
		order, err := repository.GetRechargeOrderByID(item.SourceID)
		if err == nil && order != nil {
			resp.SourceNo = order.OrderNo
			resp.SourceTitle = order.ActivityNameSnapshot
		}
	case model.RefundSourceTypeOrder:
		order, err := repository.GetOrderByID(item.SourceID)
		if err == nil && order != nil {
			resp.SourceNo = order.OrderNo
			resp.SourceTitle = fmt.Sprintf("%s订单", order.Shop.Name)
		}
	}
	return resp
}

func userRefundStatusText(status string) string {
	switch status {
	case model.RefundStatusPendingReview:
		return "待审核"
	case model.RefundStatusCancelled:
		return "已撤回"
	case model.RefundStatusProcessing:
		return "处理中"
	case model.RefundStatusSuccess:
		return "已完成"
	case model.RefundStatusRejected:
		return "已驳回"
	case model.RefundStatusFailed:
		return "处理失败"
	default:
		return ""
	}
}

func buildUserRefundResultText(item *model.RefundRequest) string {
	if item == nil {
		return ""
	}
	switch item.SourceType {
	case model.RefundSourceTypeOrder:
		if item.Status == model.RefundStatusSuccess {
			return "订单退款已退回钱包余额，可在余额流水里查看。"
		}
		return firstNonEmptyString(item.ReviewNote, item.RejectReason)
	case model.RefundSourceTypeRecharge:
		return buildUserRechargeRefundResultText(item)
	default:
		return firstNonEmptyString(item.ReviewNote, item.RejectReason)
	}
}

func CancelUserRefundRequest(userID, requestID uint) (*response.UserRefundRecordResp, error) {
	item, err := repository.GetRefundRequestByID(requestID)
	if err != nil {
		return nil, errors.New("退款申请不存在")
	}
	if item.UserID != userID {
		return nil, errors.New("无权操作该退款申请")
	}
	if !canUserCancelRefundRequest(item) {
		return nil, errors.New("当前退款申请不可撤回")
	}
	item.Status = model.RefundStatusCancelled
	if err := repository.UpdateRefundRequestWithDB(database.DB, item); err != nil {
		return nil, errors.New("撤回退款申请失败")
	}
	resp := buildUserRefundRecordResp(item)
	return &resp, nil
}

func canUserCancelRefundRequest(item *model.RefundRequest) bool {
	if item == nil {
		return false
	}
	return item.SourceType == model.RefundSourceTypeRecharge && item.RequestChannel == model.RefundRequestChannelUser && item.Status == model.RefundStatusPendingReview
}
