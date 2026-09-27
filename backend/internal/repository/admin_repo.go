package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"
)

func GetAdminByUsername(username string) (*model.Admin, error) {
	var a model.Admin
	err := database.DB.Where("username = ?", username).First(&a).Error
	return &a, err
}

func GetAdminByID(id uint) (*model.Admin, error) {
	var a model.Admin
	err := database.DB.First(&a, id).Error
	return &a, err
}

func CountMerchants() (int64, error) {
	var count int64
	err := database.DB.Model(&model.Merchant{}).Count(&count).Error
	return count, err
}

func CountShops() (int64, error) {
	var count int64
	err := database.DB.Model(&model.Shop{}).Count(&count).Error
	return count, err
}

func CountOrders() (int64, error) {
	var count int64
	err := database.DB.Model(&model.Order{}).Count(&count).Error
	return count, err
}

func CountUsers() (int64, error) {
	var count int64
	err := database.DB.Model(&model.User{}).Count(&count).Error
	return count, err
}

func SumOrderAmount() (int64, error) {
	var total int64
	err := database.DB.Model(&model.Order{}).
		Where("status IN ?", []int{model.OrderStatusPaid, model.OrderStatusAccepted, model.OrderStatusCompleted}).
		Select("COALESCE(SUM(total_amount),0)").Scan(&total).Error
	return total, err
}

func CountPendingShops() (int64, error) {
	var count int64
	err := database.DB.Model(&model.Shop{}).Where("status = ?", model.ShopStatusPending).Count(&count).Error
	return count, err
}
