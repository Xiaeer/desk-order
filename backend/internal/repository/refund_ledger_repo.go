package repository

import (
	"time"

	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetRechargeBatchByRechargeOrderID(rechargeOrderID uint) (*model.RechargeBatch, error) {
	return GetRechargeBatchByRechargeOrderIDWithDB(database.DB, rechargeOrderID)
}

func CreateRechargeBatchWithDB(db *gorm.DB, batch *model.RechargeBatch) error {
	return db.Create(batch).Error
}

func SaveRechargeBatchWithDB(db *gorm.DB, batch *model.RechargeBatch) error {
	return db.Save(batch).Error
}

func GetRechargeBatchByRechargeOrderIDWithDB(db *gorm.DB, rechargeOrderID uint) (*model.RechargeBatch, error) {
	var batch model.RechargeBatch
	err := db.Where("recharge_order_id = ?", rechargeOrderID).First(&batch).Error
	return &batch, err
}

func ListRechargeBatchesByRechargeOrderIDs(rechargeOrderIDs []uint) ([]model.RechargeBatch, error) {
	var list []model.RechargeBatch
	if len(rechargeOrderIDs) == 0 {
		return list, nil
	}
	err := database.DB.Where("recharge_order_id IN ?", rechargeOrderIDs).Find(&list).Error
	return list, err
}

func ListAvailableRechargeBatchesByUserID(userID uint) ([]model.RechargeBatch, error) {
	var list []model.RechargeBatch
	err := database.DB.
		Where("user_id = ?", userID).
		Where("principal_remaining > 0 OR gift_remaining > 0").
		Order("created_at ASC, id ASC").
		Find(&list).Error
	return list, err
}

func ListOpeningRechargeBatchesByUserID(userID uint) ([]model.RechargeBatch, error) {
	var list []model.RechargeBatch
	err := database.DB.
		Where("user_id = ?", userID).
		Where("batch_type = ?", model.RechargeBatchTypeOpening).
		Where("principal_remaining > 0 OR gift_remaining > 0").
		Order("created_at ASC, id ASC").
		Find(&list).Error
	return list, err
}

func ListAvailableRechargeBatchesByUserIDWithDB(db *gorm.DB, userID uint) ([]model.RechargeBatch, error) {
	var list []model.RechargeBatch
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		Where("principal_remaining > 0 OR gift_remaining > 0").
		Order("created_at ASC, id ASC").
		Find(&list).Error
	return list, err
}

func ListOpeningRechargeBatchesByUserIDWithDB(db *gorm.DB, userID uint) ([]model.RechargeBatch, error) {
	var list []model.RechargeBatch
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		Where("batch_type = ?", model.RechargeBatchTypeOpening).
		Where("principal_remaining > 0 OR gift_remaining > 0").
		Order("created_at ASC, id ASC").
		Find(&list).Error
	return list, err
}

func GetRechargeBatchesByIDsWithDB(db *gorm.DB, ids []uint) ([]model.RechargeBatch, error) {
	var list []model.RechargeBatch
	if len(ids) == 0 {
		return list, nil
	}
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func CreateBalanceConsumptionAllocationsWithDB(db *gorm.DB, list []model.BalanceConsumptionAllocation) error {
	if len(list) == 0 {
		return nil
	}
	return db.Create(&list).Error
}

func GetBalanceConsumptionAllocationsByOrderIDWithDB(db *gorm.DB, orderID uint) ([]model.BalanceConsumptionAllocation, error) {
	var list []model.BalanceConsumptionAllocation
	err := db.Where("order_id = ?", orderID).Order("allocation_seq ASC, id ASC").Find(&list).Error
	return list, err
}

func DeleteBalanceConsumptionAllocationsByOrderIDWithDB(db *gorm.DB, orderID uint) error {
	return db.Where("order_id = ?", orderID).Delete(&model.BalanceConsumptionAllocation{}).Error
}

func CreateLegacyOpeningRechargeBatchWithDB(db *gorm.DB, userID uint, amount int) (*model.RechargeBatch, error) {
	createdAt := time.Unix(1, 0)
	batch := &model.RechargeBatch{
		UserID:                  userID,
		BatchType:               model.RechargeBatchTypeOpening,
		ActivityNameSnapshot:    "历史余额",
		ActivityVersionSnapshot: 1,
		PrincipalTotal:          amount,
		GiftTotal:               0,
		PrincipalRemaining:      amount,
		GiftRemaining:           0,
		PrincipalConsumeWeight:  amount,
		GiftConsumeWeight:       0,
		CreatedAt:               createdAt,
		UpdatedAt:               createdAt,
	}
	if err := db.Create(batch).Error; err != nil {
		return nil, err
	}
	return batch, nil
}
