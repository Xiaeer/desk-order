package main

import (
	"fmt"
	"os"
	"strings"

	"deskorder/config"
	"deskorder/internal/model"
	"deskorder/internal/service"
	"deskorder/internal/ws"
	"deskorder/pkg/database"
	"deskorder/router"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	log.Info().Msg("config loaded")

	// 初始化 MySQL
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatal().Err(err).Msg("failed to init mysql")
	}
	log.Info().Msg("mysql connected")

	// 自动迁移表结构
	database.DB.AutoMigrate(
		&model.User{},
		&model.UserSubscribeMessageGrant{},
		&model.UserNotification{},
		&model.SystemConfig{},
		&model.UserBalance{},
		&model.RechargeBatch{},
		&model.BalanceConsumptionAllocation{},
		&model.RefundRequest{},
		&model.RefundExecution{},
		&model.BalanceTransaction{},
		&model.Merchant{},
		&model.Shop{},
		&model.ShopTable{},
		&model.ShopAuditRecord{},
		&model.Category{},
		&model.Product{},
		&model.RechargeActivity{},
		&model.RechargeOrder{},
		&model.Order{},
		&model.OrderItem{},
		&model.Admin{},
	)
	log.Info().Msg("database migrated")

	// 初始化后台管理员账号
	seedAdmin(cfg)

	// 初始化 Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatal().Err(err).Msg("failed to init redis")
	}
	log.Info().Msg("redis connected")

	// 初始化 WebSocket Hub
	ws.InitHub()
	log.Info().Msg("websocket hub started")

	// 启动退款状态自动同步任务
	service.StartRechargeRefundSyncWorker()

	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	// 设置路由
	r := router.SetupRouter()

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info().Str("addr", addr).Msg("server starting")
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

// seedAdmin 初始化后台管理员账号
func seedAdmin(cfg *config.Config) {
	if cfg == nil || !cfg.Admin.Enabled {
		log.Info().Msg("admin bootstrap disabled")
		return
	}

	username := strings.TrimSpace(cfg.Admin.Username)
	password := cfg.Admin.Password
	nickname := strings.TrimSpace(cfg.Admin.Nickname)
	if nickname == "" {
		nickname = "超级管理员"
	}

	if username == "" || strings.TrimSpace(password) == "" {
		log.Warn().Msg("admin bootstrap skipped: admin.username or admin.password is empty")
		return
	}

	var count int64
	if err := database.DB.Model(&model.Admin{}).Count(&count).Error; err != nil {
		log.Error().Err(err).Msg("failed to count admin users")
		return
	}
	if count > 0 {
		log.Info().Int64("count", count).Msg("admin account exists, skip bootstrap")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash bootstrap admin password")
		return
	}

	admin := model.Admin{
		Username: username,
		Password: string(hashed),
		Nickname: nickname,
	}
	if err := database.DB.Create(&admin).Error; err != nil {
		log.Error().Err(err).Msg("failed to create bootstrap admin account")
		return
	}

	log.Info().Str("username", username).Msg("bootstrap admin account created")
}
