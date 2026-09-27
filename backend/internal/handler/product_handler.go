package handler

import (
	"strconv"

	"deskorder/internal/dto/request"
	"deskorder/internal/repository"
	"deskorder/internal/service"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

// CreateCategory 创建菜单分类
func CreateCategory(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.CreateCategoryReq
	if !utils.BindJSON(c, &req) {
		return
	}

	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		utils.Fail(c, 404, "店铺不存在")
		return
	}

	category, err := service.CreateCategory(shop.ID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, category)
}

// UpdateCategory 更新分类
func UpdateCategory(c *gin.Context) {
	merchantID := c.GetUint("userID")
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var req request.UpdateCategoryReq
	if !utils.BindJSON(c, &req) {
		return
	}

	category, err := service.UpdateCategory(merchantID, uint(categoryID), &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, category)
}

// DeleteCategory 删除分类
func DeleteCategory(c *gin.Context) {
	merchantID := c.GetUint("userID")
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if err := service.DeleteCategory(merchantID, uint(categoryID)); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// CreateProduct 创建菜品
func CreateProduct(c *gin.Context) {
	merchantID := c.GetUint("userID")
	var req request.CreateProductReq
	if !utils.BindJSON(c, &req) {
		return
	}

	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		utils.Fail(c, 404, "店铺不存在")
		return
	}

	product, err := service.CreateProduct(shop.ID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, product)
}

// UpdateProduct 更新菜品
func UpdateProduct(c *gin.Context) {
	merchantID := c.GetUint("userID")
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var req request.UpdateProductReq
	if !utils.BindJSON(c, &req) {
		return
	}

	product, err := service.UpdateProduct(merchantID, uint(productID), &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, product)
}

// DeleteProduct 删除菜品
func DeleteProduct(c *gin.Context) {
	merchantID := c.GetUint("userID")
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if err := service.DeleteProduct(merchantID, uint(productID)); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
