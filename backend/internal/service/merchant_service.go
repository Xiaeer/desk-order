package service

import (
	"errors"
	"strings"

	"deskorder/config"
	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/auth"
	"deskorder/pkg/database"
	"deskorder/pkg/wechat"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func buildH5PlaceholderOpenID(phone string) string {
	return "h5:" + phone
}

func buildMerchantLoginResp(merchant *model.Merchant, isNew bool) (*response.MerchantLoginResp, error) {
	token, err := auth.GenerateToken(merchant.ID, "merchant")
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	return &response.MerchantLoginResp{
		Token:   token,
		IsNew:   isNew,
		HasShop: repository.MerchantHasShop(merchant.ID),
	}, nil
}

// MerchantLogin 商家微信登录
func MerchantLogin(req *request.MerchantLoginReq) (*response.MerchantLoginResp, error) {
	cfg := config.AppConfig.WeChat
	sess, err := wechat.Code2Session(cfg.MerchantAppID, cfg.MerchantSecret, req.Code)
	if err != nil {
		return nil, errors.New("微信登录失败")
	}

	merchant, err := repository.GetMerchantByOpenID(sess.OpenID)
	isNew := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		merchant = &model.Merchant{OpenID: sess.OpenID}
		if err := repository.CreateMerchant(merchant); err != nil {
			return nil, errors.New("创建商家失败")
		}
		isNew = true
	} else if err != nil {
		return nil, errors.New("查询商家失败")
	}

	return buildMerchantLoginResp(merchant, isNew)
}

// MerchantRegister 商家注册（补充信息）
func MerchantRegister(req *request.MerchantRegisterReq) (*response.MerchantLoginResp, error) {
	cfg := config.AppConfig.WeChat
	sess, err := wechat.Code2Session(cfg.MerchantAppID, cfg.MerchantSecret, req.Code)
	if err != nil {
		return nil, errors.New("微信登录失败")
	}

	merchant, err := repository.GetMerchantByOpenID(sess.OpenID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		merchant = &model.Merchant{
			OpenID: sess.OpenID,
			Name:   req.Name,
			Phone:  req.Phone,
		}
		if err := repository.CreateMerchant(merchant); err != nil {
			return nil, errors.New("注册失败")
		}
	} else if err != nil {
		return nil, errors.New("查询失败")
	} else {
		merchant.Name = req.Name
		merchant.Phone = req.Phone
		repository.UpdateMerchant(merchant)
	}

	return buildMerchantLoginResp(merchant, false)
}

// MerchantWebRegister 商户 H5 注册
func MerchantWebRegister(req *request.MerchantWebRegisterReq) (*response.MerchantLoginResp, error) {
	phone := strings.TrimSpace(req.Phone)
	name := strings.TrimSpace(req.Name)
	password := strings.TrimSpace(req.Password)

	if phone == "" || name == "" || password == "" {
		return nil, errors.New("姓名、手机号和密码不能为空")
	}

	_, err := repository.GetMerchantByPhone(phone)
	if err == nil {
		return nil, errors.New("手机号已注册")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("查询商家失败")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	merchant := &model.Merchant{
		OpenID:   buildH5PlaceholderOpenID(phone),
		Name:     name,
		Phone:    phone,
		Password: string(hashedPassword),
	}
	if err := repository.CreateMerchant(merchant); err != nil {
		return nil, errors.New("创建商户账号失败")
	}

	return buildMerchantLoginResp(merchant, false)
}

// MerchantWebLogin 商户 H5 登录
func MerchantWebLogin(req *request.MerchantWebLoginReq) (*response.MerchantLoginResp, error) {
	phone := strings.TrimSpace(req.Phone)
	password := strings.TrimSpace(req.Password)
	if phone == "" || password == "" {
		return nil, errors.New("手机号和密码不能为空")
	}

	merchant, err := repository.GetMerchantByPhone(phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("手机号或密码错误")
	}
	if err != nil {
		return nil, errors.New("查询商家失败")
	}
	if merchant.Password == "" {
		return nil, errors.New("该账号暂未设置网页登录密码")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(merchant.Password), []byte(password)); err != nil {
		return nil, errors.New("手机号或密码错误")
	}

	return buildMerchantLoginResp(merchant, false)
}

func updateMerchantWebPassword(merchantID uint, req *request.MerchantWebPasswordReq, requireCurrentPassword bool) error {
	merchant, err := repository.GetMerchantByID(merchantID)
	if err != nil {
		return errors.New("商户不存在")
	}

	newPassword := strings.TrimSpace(req.NewPassword)
	confirmPassword := strings.TrimSpace(req.ConfirmPassword)
	currentPassword := strings.TrimSpace(req.CurrentPassword)

	if newPassword == "" || confirmPassword == "" {
		return errors.New("新密码不能为空")
	}
	if len(newPassword) < 6 {
		return errors.New("新密码至少 6 位")
	}
	if newPassword != confirmPassword {
		return errors.New("两次输入的新密码不一致")
	}

	if merchant.Password != "" && requireCurrentPassword {
		if currentPassword == "" {
			return errors.New("请输入当前密码")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(merchant.Password), []byte(currentPassword)); err != nil {
			return errors.New("当前密码错误")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	merchant.Password = string(hashedPassword)
	if err := repository.UpdateMerchant(merchant); err != nil {
		return errors.New("保存密码失败")
	}

	return nil
}

// MerchantUpdateWebPassword 商户在 H5 内修改 H5 登录密码
func MerchantUpdateWebPassword(merchantID uint, req *request.MerchantWebPasswordReq) error {
	return updateMerchantWebPassword(merchantID, req, true)
}

// MerchantResetWebPasswordByMiniProgram 商户在小程序内设置或重置 H5 登录密码
func MerchantResetWebPasswordByMiniProgram(merchantID uint, req *request.MerchantWebPasswordReq) error {
	return updateMerchantWebPassword(merchantID, req, false)
}

func AdminDeleteMerchant(adminID, merchantID uint) error {
	merchant, err := repository.GetMerchantByID(merchantID)
	if err != nil {
		return errors.New("商家不存在")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		shop, shopErr := repository.GetShopByMerchantID(merchantID)
		if shopErr == nil {
			if err := deleteShopWithAudit(tx, adminID, shop); err != nil {
				return err
			}
		} else if !errors.Is(shopErr, gorm.ErrRecordNotFound) {
			return shopErr
		}

		return tx.Delete(&model.Merchant{}, merchant.ID).Error
	})
}
