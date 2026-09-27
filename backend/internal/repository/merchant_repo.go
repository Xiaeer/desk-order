package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"
)

func GetMerchantByOpenID(openID string) (*model.Merchant, error) {
	var m model.Merchant
	err := database.DB.Where("open_id = ?", openID).First(&m).Error
	return &m, err
}

func GetMerchantByPhone(phone string) (*model.Merchant, error) {
	var m model.Merchant
	err := database.DB.Where("phone = ?", phone).First(&m).Error
	return &m, err
}

func CreateMerchant(m *model.Merchant) error {
	return database.DB.Create(m).Error
}

func GetMerchantByID(id uint) (*model.Merchant, error) {
	var m model.Merchant
	err := database.DB.First(&m, id).Error
	return &m, err
}

func GetMerchantList(page, size int, status *int) ([]model.Merchant, int64, error) {
	var list []model.Merchant
	var total int64
	query := database.DB.Model(&model.Merchant{})
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func UpdateMerchant(m *model.Merchant) error {
	return database.DB.Save(m).Error
}

func DeleteMerchant(id uint) error {
	return database.DB.Delete(&model.Merchant{}, id).Error
}
