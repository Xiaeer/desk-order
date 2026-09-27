package service

import (
	"errors"

	"deskorder/internal/dto/request"
	"deskorder/internal/model"
	"deskorder/internal/repository"
)

// -------- Category --------

func CreateCategory(shopID uint, req *request.CreateCategoryReq) (*model.Category, error) {
	c := &model.Category{
		ShopID: shopID,
		Name:   req.Name,
		Sort:   req.Sort,
	}
	if err := repository.CreateCategory(c); err != nil {
		return nil, errors.New("创建分类失败")
	}
	return c, nil
}

func UpdateCategory(merchantID uint, categoryID uint, req *request.UpdateCategoryReq) (*model.Category, error) {
	c, err := repository.GetCategoryByID(categoryID)
	if err != nil {
		return nil, errors.New("分类不存在")
	}
	// 验证归属
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != c.ShopID {
		return nil, errors.New("无权操作")
	}

	if req.Name != "" {
		c.Name = req.Name
	}
	c.Sort = req.Sort

	if err := repository.UpdateCategory(c); err != nil {
		return nil, errors.New("更新失败")
	}
	return c, nil
}

func DeleteCategory(merchantID uint, categoryID uint) error {
	c, err := repository.GetCategoryByID(categoryID)
	if err != nil {
		return errors.New("分类不存在")
	}
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != c.ShopID {
		return errors.New("无权操作")
	}
	return repository.DeleteCategoryCascade(categoryID)
}

// -------- Product --------

func CreateProduct(shopID uint, req *request.CreateProductReq) (*model.Product, error) {
	// 验证分类属于该店铺
	cat, err := repository.GetCategoryByID(req.CategoryID)
	if err != nil || cat.ShopID != shopID {
		return nil, errors.New("分类不存在或不属于该店铺")
	}
	normalizedOptions, err := normalizeProductOptions(req.Options)
	if err != nil {
		return nil, err
	}

	p := &model.Product{
		ShopID:      shopID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Image:       req.Image,
		Price:       req.Price,
		Description: req.Description,
		Options:     normalizedOptions,
		IsOnSale:    true,
		Sort:        req.Sort,
	}
	if err := repository.CreateProduct(p); err != nil {
		return nil, errors.New("创建菜品失败")
	}
	return p, nil
}

func UpdateProduct(merchantID uint, productID uint, req *request.UpdateProductReq) (*model.Product, error) {
	p, err := repository.GetProductByID(productID)
	if err != nil {
		return nil, errors.New("菜品不存在")
	}
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != p.ShopID {
		return nil, errors.New("无权操作")
	}

	if req.CategoryID != 0 {
		p.CategoryID = req.CategoryID
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Image != "" {
		p.Image = req.Image
	}
	if req.Price != 0 {
		p.Price = req.Price
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Options != nil {
		normalizedOptions, err := normalizeProductOptions(*req.Options)
		if err != nil {
			return nil, err
		}
		p.Options = normalizedOptions
	}
	if req.IsOnSale != nil {
		p.IsOnSale = *req.IsOnSale
	}
	p.Sort = req.Sort

	if err := repository.UpdateProduct(p); err != nil {
		return nil, errors.New("更新失败")
	}
	return p, nil
}

func DeleteProduct(merchantID uint, productID uint) error {
	p, err := repository.GetProductByID(productID)
	if err != nil {
		return errors.New("菜品不存在")
	}
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil || shop.ID != p.ShopID {
		return errors.New("无权操作")
	}
	return repository.DeleteProduct(productID)
}
