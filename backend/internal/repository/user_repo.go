package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// GetUserByOpenID 根据 OpenID 查询用户
func GetUserByOpenID(openID string) (*model.User, error) {
	var u model.User
	err := database.DB.Where("open_id = ?", openID).First(&u).Error
	return &u, err
}

// CreateUser 创建用户
func CreateUser(u *model.User) error {
	return database.DB.Create(u).Error
}

// GetUserByID 根据 ID 查询用户
func GetUserByID(id uint) (*model.User, error) {
	var u model.User
	err := database.DB.First(&u, id).Error
	return &u, err
}

func GetUserByIDWithDB(db *gorm.DB, id uint) (*model.User, error) {
	var u model.User
	err := db.First(&u, id).Error
	return &u, err
}

func applyAdminUserKeywordFilter(query *gorm.DB, keyword, idColumn, nicknameColumn, phoneColumn, openIDColumn string) *gorm.DB {
	trimmedKeyword := strings.TrimSpace(keyword)
	if trimmedKeyword == "" {
		return query
	}
	likeKeyword := "%" + trimmedKeyword + "%"
	if userID, err := strconv.ParseUint(trimmedKeyword, 10, 64); err == nil {
		if len(trimmedKeyword) <= 6 {
			return query.Where(idColumn+" = ? OR "+nicknameColumn+" = ?", uint(userID), trimmedKeyword)
		}
		return query.Where(
			idColumn+" = ? OR "+nicknameColumn+" = ? OR "+phoneColumn+" LIKE ? OR "+openIDColumn+" LIKE ?",
			uint(userID),
			trimmedKeyword,
			likeKeyword,
			likeKeyword,
		)
	}
	return query.Where(
		nicknameColumn+" LIKE ? OR "+phoneColumn+" LIKE ? OR "+openIDColumn+" LIKE ?",
		likeKeyword,
		likeKeyword,
		likeKeyword,
	)
}

// UpdateUser 更新用户信息
func UpdateUser(u *model.User) error {
	return database.DB.Save(u).Error
}

// GetUserList 后台分页查询用户
func GetUserList(page, size int, keyword string) ([]model.User, int64, error) {
	var list []model.User
	var total int64
	query := database.DB.Model(&model.User{})
	query = applyAdminUserKeywordFilter(query, keyword, "id", "nickname", "phone", "open_id")
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetProductByIDForOrder 获取在售商品（用于下单校验）
func GetProductByIDForOrder(id uint) (*model.Product, error) {
	var p model.Product
	err := database.DB.Where("id = ? AND is_on_sale = ?", id, true).First(&p).Error
	return &p, err
}
