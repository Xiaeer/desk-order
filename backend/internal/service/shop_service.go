package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/internal/ws"
	"deskorder/pkg/database"

	"gorm.io/gorm"
)

var (
	errShopAlreadyPending  = errors.New("店铺正在审核中，请勿重复提交")
	errShopAlreadyApproved = errors.New("店铺已通过审核，无需重复申请")
	errShopAlreadyAudited  = errors.New("该店铺已审核")
	errShopHasOrders       = errors.New("店铺存在订单，不能删除")
	errShopPendingUpdate   = errors.New("店铺正在审核中，暂不可修改")
)

const shopPOSBindTokenBytes = 18

func generateShopPOSBindToken() (string, error) {
	randomBytes := make([]byte, shopPOSBindTokenBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "pos_" + hex.EncodeToString(randomBytes), nil
}

func ensureShopPOSBindToken(shop *model.Shop) error {
	if shop == nil {
		return errors.New("店铺不存在")
	}
	if shop.PosBindToken != "" {
		return nil
	}
	token, err := generateShopPOSBindToken()
	if err != nil {
		return err
	}
	shop.PosBindToken = token
	return repository.UpdateShop(shop)
}

func applyCreateShopReq(shop *model.Shop, req *request.CreateShopReq) {
	shop.Name = req.Name
	shop.Logo = req.Logo
	shop.Address = req.Address
	shop.Phone = req.Phone
	shop.Latitude = req.Latitude
	shop.Longitude = req.Longitude
	shop.Description = req.Description
}

func applyUpdateShopReq(shop *model.Shop, req *request.UpdateShopReq) bool {
	changed := false

	if req.Name != "" && req.Name != shop.Name {
		shop.Name = req.Name
		changed = true
	}
	if req.Logo != "" && req.Logo != shop.Logo {
		shop.Logo = req.Logo
		changed = true
	}
	if req.Address != "" && req.Address != shop.Address {
		shop.Address = req.Address
		changed = true
	}
	if req.Phone != "" && req.Phone != shop.Phone {
		shop.Phone = req.Phone
		changed = true
	}
	if req.Latitude != 0 && req.Latitude != shop.Latitude {
		shop.Latitude = req.Latitude
		changed = true
	}
	if req.Longitude != 0 && req.Longitude != shop.Longitude {
		shop.Longitude = req.Longitude
		changed = true
	}
	if req.Description != "" && req.Description != shop.Description {
		shop.Description = req.Description
		changed = true
	}

	return changed
}

func getMerchantOperatorName(merchantID uint) string {
	merchant, err := repository.GetMerchantByID(merchantID)
	if err != nil {
		return ""
	}
	return merchant.Name
}

func getAdminOperatorName(adminID uint) string {
	admin, err := repository.GetAdminByID(adminID)
	if err != nil {
		return ""
	}
	if admin.Nickname != "" {
		return admin.Nickname
	}
	return admin.Username
}

func newShopAuditRecord(shop *model.Shop, action, operatorRole string, operatorID uint, operatorName string, fromStatus, toStatus int, remark string) *model.ShopAuditRecord {
	return &model.ShopAuditRecord{
		ShopID:       shop.ID,
		MerchantID:   shop.MerchantID,
		Action:       action,
		OperatorRole: operatorRole,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		Remark:       remark,
	}
}

func shopAuditActionText(action string) string {
	switch action {
	case model.ShopAuditActionSubmit:
		return "首次提交入驻申请"
	case model.ShopAuditActionResubmit:
		return "重新提交入驻申请"
	case model.ShopAuditActionModify:
		return "商户提交店铺信息变更"
	case model.ShopAuditActionApprove:
		return "后台审核通过"
	case model.ShopAuditActionReject:
		return "后台审核拒绝"
	case model.ShopAuditActionDelete:
		return "后台删除店铺"
	default:
		return action
	}
}

func toAdminShopResp(shop *model.Shop) response.AdminShopResp {
	return response.AdminShopResp{
		ID:           shop.ID,
		Name:         shop.Name,
		Logo:         shop.Logo,
		Address:      shop.Address,
		Phone:        shop.Phone,
		Latitude:     shop.Latitude,
		Longitude:    shop.Longitude,
		Status:       shop.Status,
		IsOpen:       shop.IsOpen,
		Description:  shop.Description,
		MerchantID:   shop.MerchantID,
		MerchantName: shop.Merchant.Name,
		CreatedAt:    shop.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CreateShop 创建店铺
func CreateShop(merchantID uint, req *request.CreateShopReq) (*model.Shop, error) {
	existingShop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询店铺失败: %w", err)
	}

	operatorName := getMerchantOperatorName(merchantID)
	var shop *model.Shop

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			shop = &model.Shop{
				MerchantID: merchantID,
				Status:     model.ShopStatusPending,
				IsOpen:     false,
			}
			applyCreateShopReq(shop, req)
			if dbErr := tx.Create(shop).Error; dbErr != nil {
				return dbErr
			}

			record := newShopAuditRecord(
				shop,
				model.ShopAuditActionSubmit,
				model.ShopAuditOperatorMerchant,
				merchantID,
				operatorName,
				model.ShopAuditStatusNone,
				model.ShopStatusPending,
				"商户首次提交入驻申请",
			)
			return repository.CreateShopAuditRecordWithDB(tx, record)
		}

		switch existingShop.Status {
		case model.ShopStatusPending:
			return errShopAlreadyPending
		case model.ShopStatusApproved:
			return errShopAlreadyApproved
		case model.ShopStatusRejected:
			beforeStatus := existingShop.Status
			applyCreateShopReq(existingShop, req)
			existingShop.Status = model.ShopStatusPending
			existingShop.IsOpen = false
			if dbErr := tx.Save(existingShop).Error; dbErr != nil {
				return dbErr
			}

			shop = existingShop
			record := newShopAuditRecord(
				shop,
				model.ShopAuditActionResubmit,
				model.ShopAuditOperatorMerchant,
				merchantID,
				operatorName,
				beforeStatus,
				model.ShopStatusPending,
				"商户重新提交入驻申请",
			)
			return repository.CreateShopAuditRecordWithDB(tx, record)
		default:
			return errors.New("店铺状态异常，无法提交申请")
		}
	})
	if err != nil {
		return nil, err
	}
	if err := ensureShopPOSBindToken(shop); err != nil {
		return nil, errors.New("生成POS密钥失败")
	}

	return shop, nil
}

// GetMerchantShop 获取商家店铺
func GetMerchantShop(merchantID uint) (*model.Shop, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, err
	}
	if err := ensureShopPOSBindToken(shop); err != nil {
		return nil, errors.New("生成POS密钥失败")
	}
	return shop, nil
}

// UpdateShop 更新店铺信息
func UpdateShop(merchantID uint, req *request.UpdateShopReq) (*model.Shop, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}

	switch shop.Status {
	case model.ShopStatusPending:
		return nil, errShopPendingUpdate
	case model.ShopStatusRejected:
		// rejected shops can revise data and resubmit for audit
	case model.ShopStatusApproved:
		// approved shops can submit changes, but the changes must be re-audited.
	default:
		return nil, errors.New("店铺状态异常，无法修改")
	}

	beforeStatus := shop.Status
	changed := applyUpdateShopReq(shop, req)
	if beforeStatus == model.ShopStatusApproved && !changed {
		if err := ensureShopPOSBindToken(shop); err != nil {
			return nil, errors.New("生成POS密钥失败")
		}
		return shop, nil
	}

	shop.Status = model.ShopStatusPending
	shop.IsOpen = false
	operatorName := getMerchantOperatorName(merchantID)
	action := model.ShopAuditActionModify
	remark := "商户修改店铺审核信息后重新提交审核"
	if beforeStatus == model.ShopStatusRejected {
		action = model.ShopAuditActionResubmit
		remark = "商户修改店铺信息后重新提交入驻申请"
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if dbErr := tx.Save(shop).Error; dbErr != nil {
			return dbErr
		}

		record := newShopAuditRecord(
			shop,
			action,
			model.ShopAuditOperatorMerchant,
			merchantID,
			operatorName,
			beforeStatus,
			model.ShopStatusPending,
			remark,
		)
		return repository.CreateShopAuditRecordWithDB(tx, record)
	})
	if err != nil {
		return nil, fmt.Errorf("更新失败: %w", err)
	}
	if err := ensureShopPOSBindToken(shop); err != nil {
		return nil, errors.New("生成POS密钥失败")
	}

	return shop, nil
}

func RotateMerchantShopPOSBindToken(merchantID uint) (*model.Shop, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	token, err := generateShopPOSBindToken()
	if err != nil {
		return nil, errors.New("生成POS密钥失败")
	}
	shop.PosBindToken = token
	if err := repository.UpdateShop(shop); err != nil {
		return nil, errors.New("重置POS密钥失败")
	}
	if ws.DefaultHub != nil {
		ws.DefaultHub.DisconnectShop(shop.ID)
	}
	return shop, nil
}

// ToggleShopOpen 切换营业状态
func ToggleShopOpen(merchantID uint, isOpen bool) error {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusApproved {
		return errors.New("店铺尚未通过审核")
	}
	shop.IsOpen = isOpen
	return repository.UpdateShop(shop)
}

func AdminGetShopDetail(shopID uint) (*response.AdminShopResp, error) {
	shop, err := repository.GetShopByIDWithMerchant(shopID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	resp := toAdminShopResp(shop)
	return &resp, nil
}

func AdminGetShopAuditRecords(shopID uint) ([]response.ShopAuditRecordResp, error) {
	if _, err := repository.GetShopByID(shopID); err != nil {
		return nil, errors.New("店铺不存在")
	}

	records, err := repository.GetShopAuditRecords(shopID)
	if err != nil {
		return nil, errors.New("获取店铺审核记录失败")
	}

	list := make([]response.ShopAuditRecordResp, 0, len(records))
	for _, record := range records {
		list = append(list, response.ShopAuditRecordResp{
			ID:           record.ID,
			ShopID:       record.ShopID,
			MerchantID:   record.MerchantID,
			Action:       record.Action,
			ActionText:   shopAuditActionText(record.Action),
			OperatorRole: record.OperatorRole,
			OperatorID:   record.OperatorID,
			OperatorName: record.OperatorName,
			FromStatus:   record.FromStatus,
			ToStatus:     record.ToStatus,
			Remark:       record.Remark,
			CreatedAt:    record.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return list, nil
}

func AdminAuditShop(adminID, shopID uint, status int) error {
	shop, err := repository.GetShopByID(shopID)
	if err != nil {
		return errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusPending {
		return errShopAlreadyAudited
	}

	action := model.ShopAuditActionApprove
	remark := "后台审核通过"
	if status == model.ShopStatusRejected {
		action = model.ShopAuditActionReject
		remark = "后台审核拒绝"
	}

	operatorName := getAdminOperatorName(adminID)
	beforeStatus := shop.Status

	return database.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status": status,
		}
		if status == model.ShopStatusRejected {
			updates["is_open"] = false
		}

		if dbErr := tx.Model(&model.Shop{}).Where("id = ?", shopID).Updates(updates).Error; dbErr != nil {
			return dbErr
		}

		record := newShopAuditRecord(
			shop,
			action,
			model.ShopAuditOperatorAdmin,
			adminID,
			operatorName,
			beforeStatus,
			status,
			remark,
		)
		return repository.CreateShopAuditRecordWithDB(tx, record)
	})
}

func AdminDeleteShop(adminID, shopID uint) error {
	shop, err := repository.GetShopByID(shopID)
	if err != nil {
		return errors.New("店铺不存在")
	}

	return deleteShopWithAudit(database.DB, adminID, shop)
}

func deleteShopWithAudit(tx *gorm.DB, adminID uint, shop *model.Shop) error {
	if tx == nil {
		tx = database.DB
	}

	orderCount, err := repository.CountOrdersByShopID(shop.ID)
	if err != nil {
		return errors.New("检查店铺订单失败")
	}
	if orderCount > 0 {
		return errShopHasOrders
	}

	operatorName := getAdminOperatorName(adminID)
	beforeStatus := shop.Status

	return tx.Transaction(func(tx *gorm.DB) error {
		if dbErr := tx.Where("shop_id = ?", shop.ID).Delete(&model.Product{}).Error; dbErr != nil {
			return dbErr
		}
		if dbErr := tx.Where("shop_id = ?", shop.ID).Delete(&model.Category{}).Error; dbErr != nil {
			return dbErr
		}

		record := newShopAuditRecord(
			shop,
			model.ShopAuditActionDelete,
			model.ShopAuditOperatorAdmin,
			adminID,
			operatorName,
			beforeStatus,
			model.ShopAuditStatusNone,
			"后台删除店铺",
		)
		if dbErr := repository.CreateShopAuditRecordWithDB(tx, record); dbErr != nil {
			return dbErr
		}

		return tx.Delete(&model.Shop{}, shop.ID).Error
	})
}
