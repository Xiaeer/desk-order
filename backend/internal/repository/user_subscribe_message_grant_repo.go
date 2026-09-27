package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"
)

func CreateUserSubscribeMessageGrants(items []model.UserSubscribeMessageGrant) error {
	if len(items) == 0 {
		return nil
	}
	return database.DB.Create(&items).Error
}

func GetOldestPendingUserSubscribeMessageGrant(userID uint, scene string) (*model.UserSubscribeMessageGrant, error) {
	var item model.UserSubscribeMessageGrant
	err := database.DB.
		Where("user_id = ? AND scene = ? AND status = ?", userID, scene, model.SubscribeGrantStatusPending).
		Order("id ASC").
		First(&item).Error
	return &item, err
}

func UpdateUserSubscribeMessageGrant(item *model.UserSubscribeMessageGrant) error {
	return database.DB.Save(item).Error
}
