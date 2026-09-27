package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func CreateRefundRequestWithDB(db *gorm.DB, item *model.RefundRequest) error {
	return db.Create(item).Error
}

func CreateRefundExecutionWithDB(db *gorm.DB, item *model.RefundExecution) error {
	return db.Create(item).Error
}

func UpdateRefundExecutionWithDB(db *gorm.DB, item *model.RefundExecution) error {
	return db.Save(item).Error
}

func UpdateRefundRequestWithDB(db *gorm.DB, item *model.RefundRequest) error {
	return db.Save(item).Error
}

func GetRefundRequestByID(id uint) (*model.RefundRequest, error) {
	var item model.RefundRequest
	err := database.DB.First(&item, id).Error
	return &item, err
}

func GetRefundRequestByIDWithDB(db *gorm.DB, id uint) (*model.RefundRequest, error) {
	var item model.RefundRequest
	err := db.First(&item, id).Error
	return &item, err
}

func GetLatestOpenRefundRequestBySource(sourceType string, sourceID uint) (*model.RefundRequest, error) {
	return GetLatestOpenRefundRequestBySourceWithDB(database.DB, sourceType, sourceID)
}

func GetLatestOpenRefundRequestBySourceWithDB(db *gorm.DB, sourceType string, sourceID uint) (*model.RefundRequest, error) {
	var item model.RefundRequest
	err := db.
		Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Where("status IN ?", []string{model.RefundStatusPendingReview, model.RefundStatusApproved, model.RefundStatusProcessing, model.RefundStatusSuccess}).
		Order("id DESC").
		First(&item).Error
	return &item, err
}

func ListLatestRefundRequestsBySourceIDs(sourceType string, sourceIDs []uint) (map[uint]model.RefundRequest, error) {
	result := make(map[uint]model.RefundRequest)
	if len(sourceIDs) == 0 {
		return result, nil
	}
	var list []model.RefundRequest
	err := database.DB.
		Where("source_type = ?", sourceType).
		Where("source_id IN ?", sourceIDs).
		Order("source_id ASC, id DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		if _, ok := result[list[i].SourceID]; ok {
			continue
		}
		result[list[i].SourceID] = list[i]
	}
	return result, nil
}

func ListRefundRequests(page, size int, reviewScope, status, requestChannel string) ([]model.RefundRequest, int64, error) {
	var list []model.RefundRequest
	var total int64
	query := database.DB.Model(&model.RefundRequest{})
	if reviewScope != "" {
		query = query.Where("review_scope = ?", reviewScope)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if requestChannel != "" {
		query = query.Where("request_channel = ?", requestChannel)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func ListRefundRequestsByUserID(userID uint, page, size int) ([]model.RefundRequest, int64, error) {
	var list []model.RefundRequest
	var total int64
	query := database.DB.Model(&model.RefundRequest{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func CountRefundRequestsByUserID(userID uint) (int64, error) {
	var total int64
	err := database.DB.Model(&model.RefundRequest{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

func GetRefundExecutionsByRequestID(requestID uint) ([]model.RefundExecution, error) {
	var list []model.RefundExecution
	err := database.DB.Where("refund_request_id = ?", requestID).Order("id ASC").Find(&list).Error
	return list, err
}

func GetLatestRefundExecutionByRequestIDWithDB(db *gorm.DB, requestID uint) (*model.RefundExecution, error) {
	var item model.RefundExecution
	err := db.Where("refund_request_id = ?", requestID).Order("id DESC").First(&item).Error
	return &item, err
}

func GetLatestRefundExecutionByRequestID(requestID uint) (*model.RefundExecution, error) {
	return GetLatestRefundExecutionByRequestIDWithDB(database.DB, requestID)
}
