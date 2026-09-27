package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

func CreateShop(s *model.Shop) error {
	return database.DB.Create(s).Error
}

func GetShopByMerchantID(merchantID uint) (*model.Shop, error) {
	var s model.Shop
	err := database.DB.Where("merchant_id = ?", merchantID).First(&s).Error
	return &s, err
}

func GetShopByMerchantIDWithTables(merchantID uint) (*model.Shop, error) {
	var s model.Shop
	err := database.DB.
		Preload("Tables", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC, id ASC")
		}).
		Where("merchant_id = ?", merchantID).
		First(&s).Error
	return &s, err
}

func GetShopByID(id uint) (*model.Shop, error) {
	var s model.Shop
	err := database.DB.First(&s, id).Error
	return &s, err
}

func GetShopByIDWithMerchant(id uint) (*model.Shop, error) {
	var s model.Shop
	err := database.DB.Preload("Merchant").First(&s, id).Error
	return &s, err
}

func UpdateShop(s *model.Shop) error {
	return database.DB.Save(s).Error
}

func CreateShopTable(table *model.ShopTable) error {
	return database.DB.Create(table).Error
}

func UpdateShopTable(table *model.ShopTable) error {
	return database.DB.Save(table).Error
}

func DeleteShopTable(id uint) error {
	return database.DB.Delete(&model.ShopTable{}, id).Error
}

func ListShopTablesByShopID(shopID uint) ([]model.ShopTable, error) {
	var list []model.ShopTable
	err := database.DB.Where("shop_id = ?", shopID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func ListEnabledShopTablesByShopID(shopID uint) ([]model.ShopTable, error) {
	var list []model.ShopTable
	err := database.DB.
		Where("shop_id = ? AND status = ?", shopID, model.ShopTableStatusEnabled).
		Order("sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

func GetShopTableByShopIDAndID(shopID, tableID uint) (*model.ShopTable, error) {
	var table model.ShopTable
	err := database.DB.Where("shop_id = ? AND id = ?", shopID, tableID).First(&table).Error
	return &table, err
}

func GetShopTableByShopIDAndTableNo(shopID uint, tableNo string) (*model.ShopTable, error) {
	var table model.ShopTable
	err := database.DB.Where("shop_id = ? AND table_no = ?", shopID, tableNo).First(&table).Error
	return &table, err
}

func GetShopTableBySceneToken(sceneToken string) (*model.ShopTable, error) {
	var table model.ShopTable
	err := database.DB.Where("scene_token = ?", sceneToken).First(&table).Error
	return &table, err
}

func DeleteShop(id uint) error {
	return database.DB.Delete(&model.Shop{}, id).Error
}

// GetApprovedOpenShops 获取已审核且营业中的店铺
func GetApprovedOpenShops() ([]model.Shop, error) {
	var shops []model.Shop
	err := database.DB.Where("status = ? AND is_open = ?", model.ShopStatusApproved, true).Find(&shops).Error
	return shops, err
}

func GetShopList(page, size int, status *int) ([]model.Shop, int64, error) {
	var list []model.Shop
	var total int64
	query := database.DB.Model(&model.Shop{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	err := query.Preload("Merchant").Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&list).Error
	return list, total, err
}

func UpdateShopStatus(id uint, status int) error {
	return database.DB.Model(&model.Shop{}).Where("id = ?", id).Update("status", status).Error
}

func MerchantHasShop(merchantID uint) bool {
	var count int64
	database.DB.Model(&model.Shop{}).Where("merchant_id = ?", merchantID).Count(&count)
	return count > 0
}

func CreateShopAuditRecord(record *model.ShopAuditRecord) error {
	return database.DB.Create(record).Error
}

func CreateShopAuditRecordWithDB(db *gorm.DB, record *model.ShopAuditRecord) error {
	if db == nil {
		return CreateShopAuditRecord(record)
	}
	return db.Create(record).Error
}

func GetShopAuditRecords(shopID uint) ([]model.ShopAuditRecord, error) {
	var list []model.ShopAuditRecord
	err := database.DB.Where("shop_id = ?", shopID).Order("id DESC").Find(&list).Error
	return list, err
}
