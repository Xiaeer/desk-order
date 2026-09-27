package service

import (
	"errors"
	"fmt"
	"strings"

	"deskorder/internal/dto/request"
	"deskorder/internal/dto/response"
	"deskorder/internal/model"
	"deskorder/internal/repository"
	"deskorder/pkg/auth"
	"deskorder/pkg/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errAdminHistoricalBalanceNotFound     = errors.New("用户当前无可清理历史余额")
	errAdminHistoricalBalanceMultiple     = errors.New("用户存在多条历史余额批次，请先人工处理")
	errAdminHistoricalBalanceInsufficient = errors.New("用户当前余额不足，无法清理历史余额，请人工处理")
)

// AdminLogin 管理员登录
func AdminLogin(req *request.AdminLoginReq) (*response.AdminLoginResp, error) {
	admin, err := repository.GetAdminByUsername(req.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("用户名或密码错误")
	}
	if err != nil {
		return nil, errors.New("查询失败")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := auth.GenerateToken(admin.ID, "admin")
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	return &response.AdminLoginResp{
		Token:    token,
		Nickname: admin.Nickname,
	}, nil
}

// AdminGetDashboard 获取数据概览
func AdminGetDashboard() (*response.DashboardResp, error) {
	merchantCount, _ := repository.CountMerchants()
	shopCount, _ := repository.CountShops()
	orderCount, _ := repository.CountOrders()
	userCount, _ := repository.CountUsers()
	totalAmount, _ := repository.SumOrderAmount()
	pendingShops, _ := repository.CountPendingShops()

	return &response.DashboardResp{
		MerchantCount: merchantCount,
		ShopCount:     shopCount,
		OrderCount:    orderCount,
		UserCount:     userCount,
		TotalAmount:   totalAmount,
		PendingShops:  pendingShops,
	}, nil
}

func AdminGetUsers(page, size int, keyword string) ([]response.AdminUserResp, int64, error) {
	users, total, err := repository.GetUserList(page, size, keyword)
	if err != nil {
		return nil, 0, errors.New("获取用户列表失败")
	}
	list := make([]response.AdminUserResp, 0, len(users))
	for i := range users {
		item, buildErr := adminUserToResp(&users[i], false)
		if buildErr != nil {
			return nil, 0, errors.New("获取用户列表失败")
		}
		list = append(list, item)
	}
	return list, total, nil
}

func AdminGetUserDetail(userID uint) (*response.AdminUserResp, error) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	resp, err := adminUserToResp(user, true)
	if err != nil {
		return nil, errors.New("查询用户详情失败")
	}
	return &resp, nil
}

func adminUserToResp(user *model.User, includeRecentRechargeOrders bool) (response.AdminUserResp, error) {
	balance, err := repository.GetUserBalanceByUserID(user.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.AdminUserResp{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || balance == nil {
		balance = &model.UserBalance{}
	}
	orderCount, err := repository.CountOrdersByUserID(user.ID)
	if err != nil {
		return response.AdminUserResp{}, err
	}
	rechargeOrderCount, err := repository.CountRechargeOrdersByUserID(user.ID)
	if err != nil {
		return response.AdminUserResp{}, err
	}
	resp := response.AdminUserResp{
		ID:                  user.ID,
		Nickname:            user.Nickname,
		DisplayName:         buildAdminUserDisplayName(user),
		IdentityLabel:       buildAdminUserIdentityLabel(user),
		Avatar:              user.Avatar,
		Phone:               user.Phone,
		BalanceAmount:       balance.BalanceAmount,
		TotalRechargeAmount: balance.TotalRechargeAmount,
		TotalGiftAmount:     balance.TotalGiftAmount,
		TotalConsumeAmount:  balance.TotalConsumeAmount,
		OrderCount:          orderCount,
		RechargeOrderCount:  rechargeOrderCount,
		CreatedAt:           user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if !includeRecentRechargeOrders {
		return resp, nil
	}
	historicalBalanceState, err := loadAdminHistoricalBalanceState(user.ID, balance.BalanceAmount)
	if err != nil {
		return response.AdminUserResp{}, err
	}
	applyAdminHistoricalBalanceState(&resp, historicalBalanceState)
	recentRechargeOrders, err := listRecentRechargeOrdersByUserID(user.ID, 10)
	if err != nil {
		return response.AdminUserResp{}, err
	}
	resp.RecentRechargeOrders = recentRechargeOrders
	return resp, nil
}

func AdminCleanupUserHistoricalBalance(adminID, userID uint, req *request.AdminCleanupUserHistoricalBalanceReq) (*response.AdminUserResp, error) {
	cleanupRemark := strings.TrimSpace(req.Remark)
	if cleanupRemark == "" {
		cleanupRemark = "历史余额清理"
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := repository.GetUserByIDWithDB(tx, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}
		balance, err := repository.GetUserBalanceByUserIDWithDB(tx, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户余额不存在")
		}
		if err != nil {
			return err
		}
		historicalBalanceState, err := loadAdminHistoricalBalanceStateWithDB(tx, userID, balance.BalanceAmount)
		if err != nil {
			return err
		}
		if historicalBalanceState.batch == nil {
			if strings.TrimSpace(historicalBalanceState.unavailableReason) != "" {
				return errors.New(historicalBalanceState.unavailableReason)
			}
			return errAdminHistoricalBalanceNotFound
		}
		if historicalBalanceState.amount <= 0 {
			return errAdminHistoricalBalanceNotFound
		}
		if balance.BalanceAmount < historicalBalanceState.amount {
			return errAdminHistoricalBalanceInsufficient
		}

		balance.BalanceAmount -= historicalBalanceState.amount
		if err := repository.SaveUserBalanceWithDB(tx, balance); err != nil {
			return err
		}

		historicalBalanceState.batch.PrincipalRemaining = 0
		historicalBalanceState.batch.GiftRemaining = 0
		if err := repository.SaveRechargeBatchWithDB(tx, historicalBalanceState.batch); err != nil {
			return err
		}

		remark := buildAdminHistoricalBalanceCleanupRemark(adminID, userID, historicalBalanceState.batch.ID, cleanupRemark)
		return repository.CreateBalanceTransactionWithDB(tx, &model.BalanceTransaction{
			UserID:                userID,
			ChangeAmount:          -historicalBalanceState.amount,
			PrincipalChangeAmount: 0,
			GiftChangeAmount:      0,
			BalanceAfter:          balance.BalanceAmount,
			BizType:               model.BalanceTransactionTypeAdjust,
			SourceType:            model.BalanceTransactionSourceAdminAdjust,
			SourceID:              adminID,
			RelatedBatchID:        &historicalBalanceState.batch.ID,
			Remark:                remark,
		})
	})
	if err != nil {
		switch err.Error() {
		case "用户不存在", "用户余额不存在", errAdminHistoricalBalanceNotFound.Error(), errAdminHistoricalBalanceMultiple.Error(), errAdminHistoricalBalanceInsufficient.Error():
			return nil, err
		default:
			return nil, errors.New("清理用户历史余额失败")
		}
	}
	return AdminGetUserDetail(userID)
}

func buildAdminUserDisplayName(user *model.User) string {
	if nickname := strings.TrimSpace(user.Nickname); nickname != "" {
		return nickname
	}
	return fmt.Sprintf("用户#%d", user.ID)
}

func buildAdminUserIdentityLabel(user *model.User) string {
	if phone := strings.TrimSpace(user.Phone); phone != "" {
		return phone
	}
	if suffix := buildAdminUserOpenIDTail(user.OpenID); suffix != "" {
		return "OpenID尾号 " + suffix
	}
	return fmt.Sprintf("用户ID %d", user.ID)
}

func buildAdminUserOpenIDTail(openID string) string {
	trimmed := strings.TrimSpace(openID)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 6 {
		return trimmed
	}
	return trimmed[len(trimmed)-6:]
}

type adminHistoricalBalanceState struct {
	amount            int
	batch             *model.RechargeBatch
	unavailableReason string
}

func loadAdminHistoricalBalanceState(userID uint, currentBalance int) (*adminHistoricalBalanceState, error) {
	batches, err := repository.ListOpeningRechargeBatchesByUserID(userID)
	if err != nil {
		return nil, err
	}
	return buildAdminHistoricalBalanceState(batches, currentBalance), nil
}

func loadAdminHistoricalBalanceStateWithDB(db *gorm.DB, userID uint, currentBalance int) (*adminHistoricalBalanceState, error) {
	batches, err := repository.ListOpeningRechargeBatchesByUserIDWithDB(db, userID)
	if err != nil {
		return nil, err
	}
	return buildAdminHistoricalBalanceState(batches, currentBalance), nil
}

func buildAdminHistoricalBalanceState(batches []model.RechargeBatch, currentBalance int) *adminHistoricalBalanceState {
	state := &adminHistoricalBalanceState{}
	if len(batches) == 0 {
		return state
	}
	for i := range batches {
		state.amount += batches[i].PrincipalRemaining + batches[i].GiftRemaining
	}
	if len(batches) > 1 {
		state.unavailableReason = errAdminHistoricalBalanceMultiple.Error()
		return state
	}
	batch := batches[0]
	state.batch = &batch
	if state.amount <= 0 {
		state.unavailableReason = errAdminHistoricalBalanceNotFound.Error()
		return state
	}
	if currentBalance < state.amount {
		state.unavailableReason = errAdminHistoricalBalanceInsufficient.Error()
	}
	return state
}

func applyAdminHistoricalBalanceState(resp *response.AdminUserResp, state *adminHistoricalBalanceState) {
	if resp == nil || state == nil {
		return
	}
	resp.HistoricalBalanceAmount = state.amount
	resp.HistoricalBalanceUnavailableReason = state.unavailableReason
	if state.batch == nil {
		resp.CanCleanupHistoricalBalance = false
		return
	}
	resp.HistoricalBalanceBatchID = state.batch.ID
	resp.CanCleanupHistoricalBalance = state.amount > 0 && strings.TrimSpace(state.unavailableReason) == ""
}

func buildAdminHistoricalBalanceCleanupRemark(adminID, userID, batchID uint, remark string) string {
	base := "后台历史余额清理"
	if trimmedRemark := strings.TrimSpace(remark); trimmedRemark != "" {
		base += "：" + trimmedRemark
	}
	return fmt.Sprintf("%s（用户#%d，历史批次#%d，管理员#%d）", base, userID, batchID, adminID)
}

func listRecentRechargeOrdersByUserID(userID uint, limit int) ([]response.AdminRechargeOrderResp, error) {
	orders, err := repository.ListRecentAdminRechargeOrdersByUserID(userID, limit)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	orderIDs := make([]uint, 0, len(orders))
	for i := range orders {
		orderIDs = append(orderIDs, orders[i].ID)
	}
	refundMap, err := repository.ListLatestRefundRequestsBySourceIDs(model.RefundSourceTypeRecharge, orderIDs)
	if err != nil {
		return nil, err
	}
	list := make([]response.AdminRechargeOrderResp, 0, len(orders))
	for i := range orders {
		list = append(list, adminRechargeOrderToResp(&orders[i], refundMap[orders[i].ID], nil))
	}
	return list, nil
}
