package repository

import (
	"strings"
	"time"

	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func GetUserBalanceByUserID(userID uint) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := database.DB.Where("user_id = ?", userID).First(&balance).Error
	return &balance, err
}

func GetOrCreateUserBalanceByUserID(userID uint) (*model.UserBalance, error) {
	balance, err := GetUserBalanceByUserID(userID)
	if err == nil {
		return balance, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	balance = &model.UserBalance{UserID: userID}
	if err := database.DB.Create(balance).Error; err != nil {
		return nil, err
	}
	return balance, nil
}

func GetUserBalanceByUserIDWithDB(db *gorm.DB, userID uint) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := db.Where("user_id = ?", userID).First(&balance).Error
	return &balance, err
}

func SaveUserBalanceWithDB(db *gorm.DB, balance *model.UserBalance) error {
	return db.Save(balance).Error
}

func CreateBalanceTransactionWithDB(db *gorm.DB, record *model.BalanceTransaction) error {
	return db.Create(record).Error
}

func GetBalanceTransactionsByUserID(userID uint, page, size int) ([]model.BalanceTransaction, int64, error) {
	var list []model.BalanceTransaction
	var total int64
	query := database.DB.Model(&model.BalanceTransaction{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Preload("Shop").Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func GetActiveRechargeActivities(now time.Time) ([]model.RechargeActivity, error) {
	var list []model.RechargeActivity
	err := database.DB.
		Where("status = ?", model.RechargeActivityStatusEnabled).
		Where("start_at IS NULL OR start_at <= ?", now).
		Where("end_at IS NULL OR end_at >= ?", now).
		Order("sort ASC, recharge_amount ASC, id ASC").
		Find(&list).Error
	return list, err
}

func GetRechargeActivities(page, size int, status *int) ([]model.RechargeActivity, int64, error) {
	var list []model.RechargeActivity
	var total int64
	query := database.DB.Model(&model.RechargeActivity{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("sort ASC, recharge_amount ASC, id ASC").Find(&list).Error
	return list, total, err
}

func GetRechargeActivityByID(id uint) (*model.RechargeActivity, error) {
	var activity model.RechargeActivity
	err := database.DB.First(&activity, id).Error
	return &activity, err
}

func CountRechargeOrdersByActivityID(activityID uint) (int64, error) {
	var total int64
	err := database.DB.Model(&model.RechargeOrder{}).Where("activity_id = ?", activityID).Count(&total).Error
	return total, err
}

func CountRechargeOrdersByUserID(userID uint) (int64, error) {
	var total int64
	err := database.DB.Model(&model.RechargeOrder{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

func CreateRechargeActivity(activity *model.RechargeActivity) error {
	return database.DB.Create(activity).Error
}

func UpdateRechargeActivity(activity *model.RechargeActivity) error {
	return database.DB.Save(activity).Error
}

func DeleteRechargeActivity(id uint) error {
	return database.DB.Delete(&model.RechargeActivity{}, id).Error
}

func CreateRechargeOrder(order *model.RechargeOrder) error {
	return database.DB.Create(order).Error
}

func GetRechargeOrdersByUserID(userID uint, page, size int, status *int) ([]model.RechargeOrder, int64, error) {
	var list []model.RechargeOrder
	var total int64
	query := database.DB.Model(&model.RechargeOrder{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	err := query.Preload("Activity").Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func ListPaidRechargeOrdersWithoutBatchByUserID(userID uint) ([]model.RechargeOrder, error) {
	return ListPaidRechargeOrdersWithoutBatchByUserIDWithDB(database.DB, userID)
}

func ListPaidRechargeOrdersWithoutBatchByUserIDWithDB(db *gorm.DB, userID uint) ([]model.RechargeOrder, error) {
	var list []model.RechargeOrder
	err := db.Model(&model.RechargeOrder{}).
		Joins("LEFT JOIN recharge_batches ON recharge_batches.recharge_order_id = recharge_orders.id").
		Where("recharge_orders.user_id = ?", userID).
		Where("recharge_orders.status = ?", model.RechargeOrderStatusPaid).
		Where("recharge_batches.id IS NULL").
		Order("recharge_orders.id ASC").
		Find(&list).Error
	return list, err
}

func ListRecentAdminRechargeOrdersByUserID(userID uint, limit int) ([]model.RechargeOrder, error) {
	var list []model.RechargeOrder
	query := database.DB.Where("user_id = ?", userID).
		Preload("Activity").
		Preload("User").
		Order("id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&list).Error
	return list, err
}

func GetAdminRechargeOrders(page, size int, status *int, orderNo, userKeyword string, createdStartAt, createdEndAt *time.Time) ([]model.RechargeOrder, int64, error) {
	query := buildAdminRechargeOrdersQuery(status, orderNo, userKeyword, createdStartAt, createdEndAt)
	var list []model.RechargeOrder
	var total int64
	query.Count(&total)
	err := query.
		Preload("Activity").
		Preload("User").
		Offset((page - 1) * size).
		Limit(size).
		Order("recharge_orders.id DESC").
		Find(&list).Error
	return list, total, err
}

func ListAdminRechargeOrders(status *int, orderNo, userKeyword string, createdStartAt, createdEndAt *time.Time) ([]model.RechargeOrder, error) {
	var list []model.RechargeOrder
	err := buildAdminRechargeOrdersQuery(status, orderNo, userKeyword, createdStartAt, createdEndAt).
		Preload("Activity").
		Preload("User").
		Order("recharge_orders.id DESC").
		Find(&list).Error
	return list, err
}

func buildAdminRechargeOrdersQuery(status *int, orderNo, userKeyword string, createdStartAt, createdEndAt *time.Time) *gorm.DB {
	query := database.DB.Model(&model.RechargeOrder{}).
		Joins("LEFT JOIN users ON users.id = recharge_orders.user_id")
	if status != nil {
		query = query.Where("recharge_orders.status = ?", *status)
	}
	if trimmedOrderNo := strings.TrimSpace(orderNo); trimmedOrderNo != "" {
		query = query.Where("recharge_orders.order_no LIKE ?", "%"+trimmedOrderNo+"%")
	}
	query = applyAdminUserKeywordFilter(query, userKeyword, "users.id", "users.nickname", "users.phone", "users.open_id")
	if createdStartAt != nil {
		query = query.Where("recharge_orders.created_at >= ?", *createdStartAt)
	}
	if createdEndAt != nil {
		query = query.Where("recharge_orders.created_at <= ?", *createdEndAt)
	}
	return query
}

func GetRechargeOrderByID(id uint) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := database.DB.Preload("Activity").First(&order, id).Error
	return &order, err
}

func GetRechargeOrderByIDWithDB(db *gorm.DB, id uint) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := db.Preload("Activity").First(&order, id).Error
	return &order, err
}

func GetAdminRechargeOrderByID(id uint) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := database.DB.Preload("Activity").Preload("User").First(&order, id).Error
	return &order, err
}

func GetRechargeOrderByUserIDAndID(userID, id uint) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := database.DB.Preload("Activity").Where("user_id = ? AND id = ?", userID, id).First(&order).Error
	return &order, err
}

func GetRechargeOrderByOrderNo(orderNo string) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := database.DB.Preload("Activity").Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

func UpdateRechargeOrder(order *model.RechargeOrder) error {
	return database.DB.Save(order).Error
}
