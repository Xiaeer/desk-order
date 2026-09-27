package repository

import (
	"deskorder/internal/model"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

// -------- Category --------

func CreateCategory(c *model.Category) error {
	return database.DB.Create(c).Error
}

func GetCategoriesByShopID(shopID uint) ([]model.Category, error) {
	var list []model.Category
	err := database.DB.Where("shop_id = ?", shopID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func GetCategoryByID(id uint) (*model.Category, error) {
	var c model.Category
	err := database.DB.First(&c, id).Error
	return &c, err
}

func UpdateCategory(c *model.Category) error {
	return database.DB.Save(c).Error
}

func DeleteCategory(id uint) error {
	return database.DB.Delete(&model.Category{}, id).Error
}

func DeleteCategoryCascade(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("category_id = ?", id).Delete(&model.Product{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Category{}, id).Error
	})
}

// -------- Product --------

func CreateProduct(p *model.Product) error {
	return database.DB.Create(p).Error
}

func GetProductsByShopID(shopID uint) ([]model.Product, error) {
	var list []model.Product
	err := database.DB.Where("shop_id = ?", shopID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func GetProductsByCategoryID(categoryID uint) ([]model.Product, error) {
	var list []model.Product
	err := database.DB.Where("category_id = ?", categoryID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func GetProductByID(id uint) (*model.Product, error) {
	var p model.Product
	err := database.DB.First(&p, id).Error
	return &p, err
}

func UpdateProduct(p *model.Product) error {
	return database.DB.Save(p).Error
}

func DeleteProduct(id uint) error {
	return database.DB.Delete(&model.Product{}, id).Error
}

func DeleteProductsByShopID(shopID uint) error {
	return database.DB.Where("shop_id = ?", shopID).Delete(&model.Product{}).Error
}

func DeleteCategoriesByShopID(shopID uint) error {
	return database.DB.Where("shop_id = ?", shopID).Delete(&model.Category{}).Error
}

// GetOnSaleProductsByCategoryID 获取某分类下在售菜品
func GetOnSaleProductsByCategoryID(categoryID uint) ([]model.Product, error) {
	var list []model.Product
	err := database.DB.Where("category_id = ? AND is_on_sale = ?", categoryID, true).
		Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}
