package repository

import (
	"time"

	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func CreateUserNotificationWithDB(db *gorm.DB, item *model.UserNotification) error {
	return db.Create(item).Error
}

func CountUnreadUserNotificationsByCategory(userID uint, category string) (int64, error) {
	var total int64
	err := database.DB.Model(&model.UserNotification{}).
		Where("user_id = ? AND category = ? AND is_read = ?", userID, category, false).
		Count(&total).Error
	return total, err
}

func ListUserNotificationsByCategory(userID uint, category string, page, size int) ([]model.UserNotification, int64, error) {
	var list []model.UserNotification
	var total int64
	query := database.DB.Model(&model.UserNotification{}).
		Where("user_id = ? AND category = ?", userID, category)
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func MarkAllUserNotificationsReadByCategory(userID uint, category string) error {
	now := time.Now()
	return database.DB.Model(&model.UserNotification{}).
		Where("user_id = ? AND category = ? AND is_read = ?", userID, category, false).
		Updates(map[string]any{"is_read": true, "read_at": &now}).Error
}
