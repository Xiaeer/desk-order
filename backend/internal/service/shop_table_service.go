package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"deskorder/config"
	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/wechat"

	"gorm.io/gorm"
)

const (
	shopTableQRCodeDir       = "uploads/shop-tables"
	shopTableEntryPage       = "pages/index/index"
	shopTableTokenBytes      = 12
	miniAppEnvVersionDevelop = "develop"
	miniAppEnvVersionTrial   = "trial"
	miniAppEnvVersionRelease = "release"
)

func normalizeShopTableNo(raw string) string {
	return strings.TrimSpace(raw)
}

func toShopTableResp(table *model.ShopTable) response.ShopTableResp {
	if table == nil {
		return response.ShopTableResp{}
	}
	return response.ShopTableResp{
		ID:         table.ID,
		ShopID:     table.ShopID,
		TableNo:    table.TableNo,
		SceneToken: table.SceneToken,
		QRCodeURL:  table.QRCodeURL,
		Status:     table.Status,
		Sort:       table.Sort,
	}
}

func generateShopTableSceneToken() (string, error) {
	randomBytes := make([]byte, shopTableTokenBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "st" + hex.EncodeToString(randomBytes), nil
}

func getMiniAppQRCodeEnvVersion() (string, error) {
	if config.AppConfig == nil {
		return "", errors.New("系统配置未加载")
	}
	envVersion := strings.ToLower(strings.TrimSpace(config.AppConfig.WeChat.UserCodeEnvVersion))
	if envVersion == "" {
		if strings.EqualFold(strings.TrimSpace(config.AppConfig.Server.Mode), "release") {
			return miniAppEnvVersionRelease, nil
		}
		return miniAppEnvVersionDevelop, nil
	}
	switch envVersion {
	case miniAppEnvVersionDevelop, miniAppEnvVersionTrial, miniAppEnvVersionRelease:
		return envVersion, nil
	default:
		return "", fmt.Errorf("invalid mini app env version: %s", envVersion)
	}
}

func buildShopTableQRCodeFilename(sceneToken, envVersion string) string {
	return fmt.Sprintf("%s-%s.png", sceneToken, envVersion)
}

func buildShopTableQRCodeURL(sceneToken, envVersion string) string {
	return "/uploads/shop-tables/" + buildShopTableQRCodeFilename(sceneToken, envVersion)
}

func buildShopTableQRCodePath(sceneToken, envVersion string) string {
	return filepath.Join(shopTableQRCodeDir, buildShopTableQRCodeFilename(sceneToken, envVersion))
}

func generateShopTableQRCode(sceneToken string) (string, error) {
	if config.AppConfig == nil {
		return "", errors.New("系统配置未加载")
	}
	envVersion, err := getMiniAppQRCodeEnvVersion()
	if err != nil {
		return "", err
	}
	wechatCfg := config.AppConfig.WeChat
	qrcodeBytes, err := wechat.GetMiniAppUnlimitedQRCode(&wechat.UnlimitedQRCodeReq{
		AppID:      wechatCfg.UserAppID,
		Secret:     wechatCfg.UserSecret,
		Scene:      sceneToken,
		Page:       shopTableEntryPage,
		EnvVersion: envVersion,
		CheckPath:  false,
		Width:      430,
	})
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(shopTableQRCodeDir, 0o755); err != nil {
		return "", err
	}
	savePath := buildShopTableQRCodePath(sceneToken, envVersion)
	if err := os.WriteFile(savePath, qrcodeBytes, 0o644); err != nil {
		return "", err
	}
	return buildShopTableQRCodeURL(sceneToken, envVersion), nil
}

func shopTableQRCodeFileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func ensureShopTableQRCode(table *model.ShopTable) error {
	if table == nil {
		return errors.New("桌号不存在")
	}
	envVersion, err := getMiniAppQRCodeEnvVersion()
	if err != nil {
		return err
	}
	expectedURL := buildShopTableQRCodeURL(table.SceneToken, envVersion)
	expectedPath := buildShopTableQRCodePath(table.SceneToken, envVersion)
	if table.QRCodeURL == expectedURL && shopTableQRCodeFileExists(expectedPath) {
		return nil
	}
	qrcodeURL, err := generateShopTableQRCode(table.SceneToken)
	if err != nil {
		return err
	}
	table.QRCodeURL = qrcodeURL
	return repository.UpdateShopTable(table)
}

func ListMerchantShopTables(merchantID uint) ([]response.ShopTableResp, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	list, err := repository.ListShopTablesByShopID(shop.ID)
	if err != nil {
		return nil, errors.New("获取桌号列表失败")
	}
	resp := make([]response.ShopTableResp, 0, len(list))
	for i := range list {
		if err := ensureShopTableQRCode(&list[i]); err != nil {
			return nil, errors.New("刷新桌码失败")
		}
		resp = append(resp, toShopTableResp(&list[i]))
	}
	return resp, nil
}

func CreateMerchantShopTable(merchantID uint, req *request.CreateShopTableReq) (*response.ShopTableResp, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	tableNo := normalizeShopTableNo(req.TableNo)
	if tableNo == "" {
		return nil, errors.New("桌号不能为空")
	}
	if _, err := repository.GetShopTableByShopIDAndTableNo(shop.ID, tableNo); err == nil {
		return nil, errors.New("桌号已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("查询桌号失败")
	}
	sceneToken, err := generateShopTableSceneToken()
	if err != nil {
		return nil, errors.New("生成桌码失败")
	}
	qrcodeURL, err := generateShopTableQRCode(sceneToken)
	if err != nil {
		return nil, errors.New("生成桌码失败")
	}
	table := &model.ShopTable{
		ShopID:     shop.ID,
		TableNo:    tableNo,
		SceneToken: sceneToken,
		QRCodeURL:  qrcodeURL,
		Status:     model.ShopTableStatusEnabled,
		Sort:       req.Sort,
	}
	if err := repository.CreateShopTable(table); err != nil {
		return nil, errors.New("创建桌号失败")
	}
	resp := toShopTableResp(table)
	return &resp, nil
}

func UpdateMerchantShopTable(merchantID, tableID uint, req *request.UpdateShopTableReq) (*response.ShopTableResp, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	table, err := repository.GetShopTableByShopIDAndID(shop.ID, tableID)
	if err != nil {
		return nil, errors.New("桌号不存在")
	}
	if req.TableNo != nil {
		tableNo := normalizeShopTableNo(*req.TableNo)
		if tableNo == "" {
			return nil, errors.New("桌号不能为空")
		}
		if tableNo != table.TableNo {
			if existing, findErr := repository.GetShopTableByShopIDAndTableNo(shop.ID, tableNo); findErr == nil && existing.ID != table.ID {
				return nil, errors.New("桌号已存在")
			} else if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return nil, errors.New("查询桌号失败")
			}
			table.TableNo = tableNo
		}
	}
	if req.Status != nil {
		if *req.Status != model.ShopTableStatusEnabled && *req.Status != model.ShopTableStatusDisabled {
			return nil, errors.New("桌号状态无效")
		}
		table.Status = *req.Status
	}
	if req.Sort != nil {
		table.Sort = *req.Sort
	}
	if err := repository.UpdateShopTable(table); err != nil {
		return nil, errors.New("更新桌号失败")
	}
	resp := toShopTableResp(table)
	return &resp, nil
}

func DeleteMerchantShopTable(merchantID, tableID uint) error {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return errors.New("店铺不存在")
	}
	table, err := repository.GetShopTableByShopIDAndID(shop.ID, tableID)
	if err != nil {
		return errors.New("桌号不存在")
	}
	if err := repository.DeleteShopTable(table.ID); err != nil {
		return errors.New("删除桌号失败")
	}
	return nil
}

func ToggleMerchantShopAutoAcceptOrders(merchantID uint, enabled bool) (*model.Shop, error) {
	shop, err := repository.GetShopByMerchantID(merchantID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusApproved {
		return nil, errors.New("店铺尚未通过审核")
	}
	shop.AutoAcceptOrders = enabled
	if err := repository.UpdateShop(shop); err != nil {
		return nil, errors.New("更新自动接单设置失败")
	}
	if err := ensureShopPOSBindToken(shop); err != nil {
		return nil, errors.New("生成POS密钥失败")
	}
	return shop, nil
}

func ListUserShopTables(shopID uint) ([]response.ShopTableResp, error) {
	shop, err := repository.GetShopByID(shopID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusApproved || !shop.IsOpen {
		return nil, errors.New("店铺未营业")
	}
	list, err := repository.ListEnabledShopTablesByShopID(shop.ID)
	if err != nil {
		return nil, errors.New("获取桌号列表失败")
	}
	resp := make([]response.ShopTableResp, 0, len(list))
	for i := range list {
		resp = append(resp, toShopTableResp(&list[i]))
	}
	return resp, nil
}

func ResolveUserShopTableScene(sceneToken string) (*response.ShopTableSceneResp, error) {
	token := strings.TrimSpace(sceneToken)
	if token == "" {
		return nil, errors.New("桌码不能为空")
	}
	table, err := repository.GetShopTableBySceneToken(token)
	if err != nil {
		return nil, errors.New("桌码不存在")
	}
	if table.Status != model.ShopTableStatusEnabled {
		return nil, errors.New("该桌号已停用")
	}
	shop, err := repository.GetShopByID(table.ShopID)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	if shop.Status != model.ShopStatusApproved || !shop.IsOpen {
		return nil, errors.New("店铺未营业")
	}
	resp := &response.ShopTableSceneResp{
		ShopID:      shop.ID,
		ShopName:    shop.Name,
		ShopTableID: table.ID,
		TableNo:     table.TableNo,
		SceneToken:  table.SceneToken,
	}
	return resp, nil
}

func resolveCreateOrderShopTable(shopID uint, shopTableID *uint, entryScene string) (*model.ShopTable, error) {
	sceneToken := strings.TrimSpace(entryScene)
	if shopTableID == nil {
		if sceneToken == "" {
			return nil, nil
		}
		table, err := repository.GetShopTableBySceneToken(sceneToken)
		if err != nil {
			return nil, errors.New("桌码不存在")
		}
		if table.ShopID != shopID {
			return nil, errors.New("桌码与店铺不匹配")
		}
		if table.Status != model.ShopTableStatusEnabled {
			return nil, errors.New("该桌号已停用")
		}
		return table, nil
	}
	table, err := repository.GetShopTableByShopIDAndID(shopID, *shopTableID)
	if err != nil {
		return nil, errors.New("桌号不存在")
	}
	if table.Status != model.ShopTableStatusEnabled {
		return nil, errors.New("该桌号已停用")
	}
	if sceneToken != "" && sceneToken != table.SceneToken {
		return nil, errors.New("桌码与桌号不匹配")
	}
	return table, nil
}
