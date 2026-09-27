package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func CreateOrder(o *model.Order) error {
	return database.DB.Create(o).Error
}

func GetOrderByID(id uint) (*model.Order, error) {
	var o model.Order
	err := database.DB.Preload("Items").Preload("Shop").First(&o, id).Error
	return &o, err
}

func GetOrderByIDWithDB(db *gorm.DB, id uint) (*model.Order, error) {
	var o model.Order
	err := db.Preload("Items").Preload("Shop").First(&o, id).Error
	return &o, err
}

func GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var o model.Order
	err := database.DB.Preload("Items").Preload("Shop").Where("order_no = ?", orderNo).First(&o).Error
	return &o, err
}

func GetOrderByShopIDAndClientRequestID(shopID uint, clientRequestID string) (*model.Order, error) {
	var o model.Order
	err := database.DB.Preload("Items").Preload("Shop").Where("shop_id = ? AND client_request_id = ?", shopID, clientRequestID).First(&o).Error
	return &o, err
}

func UpdateOrder(o *model.Order) error {
	return database.DB.Save(o).Error
}

func UpdateOrderWithDB(db *gorm.DB, o *model.Order) error {
	return db.Save(o).Error
}

func GetOrdersByUserID(userID uint, page, size int) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	query := database.DB.Model(&model.Order{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Preload("Items").Preload("Shop").
		Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func GetOrdersByShopID(shopID uint, status *int, page, size int) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	query := database.DB.Model(&model.Order{}).Where("shop_id = ?", shopID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	err := query.Preload("Items").
		Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func GetAllOrders(page, size int, status *int) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	query := database.DB.Model(&model.Order{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	err := query.Preload("Items").Preload("Shop").
		Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func CountOrdersByShopID(shopID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Order{}).Where("shop_id = ?", shopID).Count(&count).Error
	return count, err
}

func CountOrdersByUserID(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Order{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func UpdateOrderStatus(id uint, status int) error {
	return database.DB.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
