package repository

import (
	"errors"

	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func GetSystemConfigsByKeys(keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	if len(keys) == 0 {
		return values, nil
	}

	var list []model.SystemConfig
	if err := database.DB.Where("config_key IN ?", keys).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, item := range list {
		values[item.ConfigKey] = item.ConfigValue
	}
	return values, nil
}

func GetSystemConfigByKey(key string) (*model.SystemConfig, error) {
	var item model.SystemConfig
	err := database.DB.Where("config_key = ?", key).First(&item).Error
	return &item, err
}

func UpsertSystemConfig(item *model.SystemConfig) error {
	existing, err := GetSystemConfigByKey(item.ConfigKey)
	if err == nil {
		existing.ConfigValue = item.ConfigValue
		existing.ValueType = item.ValueType
		existing.Description = item.Description
		existing.UpdatedBy = item.UpdatedBy
		return database.DB.Save(existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return database.DB.Create(item).Error
}
