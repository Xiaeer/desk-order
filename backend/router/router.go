package router

import (
	"deskorder/internal/handler"
	"deskorder/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Cors(), gin.Recovery())
	r.Static("/uploads", "./uploads")

	// --------------- 健康检查 ---------------
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "pong"})
	})

	api := r.Group("/api/v1")

	// --------------- 用户端（小程序） ---------------
	userGroup := api.Group("/user")
	{
		userGroup.POST("/login", handler.UserLogin) // 微信登录
		userGroup.GET("/config/public", handler.GetMiniUserPublicConfig)
		userGroup.GET("/shops/nearby", handler.GetNearbyShops) // 附近商家列表
		userGroup.GET("/shop/:id/menu", handler.GetShopMenu)   // 商家菜单
		userGroup.GET("/shop/:id/tables", handler.GetUserShopTables)
		userGroup.GET("/entry/scene", handler.ResolveShopTableScene)

		// 需要登录
		userAuth := userGroup.Group("", middleware.JWTAuth(), middleware.RoleAuth("user"))
		{
			userAuth.POST("/upload/avatar", handler.UploadUserAvatar)
			userAuth.GET("/me", handler.GetUserProfile)
			userAuth.PUT("/me/profile", handler.UpdateUserProfile)
			userAuth.POST("/me/phone", handler.BindUserPhone)
			userAuth.GET("/recharge/activities", handler.GetRechargeActivities)
			userAuth.GET("/recharge/orders", handler.GetUserRechargeOrders)
			userAuth.GET("/refund-records", handler.GetUserRefundRecords)
			userAuth.GET("/refund-notifications", handler.GetUserRefundNotifications)
			userAuth.PUT("/refund-notifications/read", handler.ReadAllUserRefundNotifications)
			userAuth.POST("/refund-request/:id/cancel", handler.CancelUserRefundRequest)
			userAuth.POST("/refund-subscriptions/grant", handler.GrantUserRefundSubscribePermissions)
			userAuth.POST("/recharge/order", handler.CreateRechargeOrder)
			userAuth.POST("/recharge/order/:id/refund", handler.ApplyUserRechargeRefund)
			userAuth.DELETE("/recharge/order/:id", handler.CancelUserRechargeOrder)
			userAuth.POST("/recharge/order/:id/pay", handler.CreateUserRechargePay)
			userAuth.GET("/balance/transactions", handler.GetBalanceTransactions)
			userAuth.POST("/order", handler.CreateOrder) // 下单
			userAuth.POST("/order/:id/pay", handler.CreateUserOrderPay)
			userAuth.POST("/order/:id/cancel", handler.CancelUserOrder)
			userAuth.GET("/orders", handler.GetUserOrders)         // 我的订单
			userAuth.GET("/order/:id", handler.GetUserOrderDetail) // 订单详情
		}
	}

	// --------------- 支付回调 ---------------
	api.POST("/pay/notify", handler.PayNotify) // 微信支付回调

	// --------------- 商家端（小程序） ---------------
	merchantGroup := api.Group("/merchant")
	{
		merchantGroup.POST("/login", handler.MerchantLogin)       // 商家登录
		merchantGroup.POST("/register", handler.MerchantRegister) // 商家注册
		merchantGroup.POST("/web/login", handler.MerchantWebLogin)
		merchantGroup.POST("/web/register", handler.MerchantWebRegister)
		merchantGroup.GET("/ws", handler.MerchantWebSocket)

		// 需要商家登录
		mAuth := merchantGroup.Group("", middleware.JWTAuth(), middleware.RoleAuth("merchant"))
		{
			mAuth.POST("/upload/product-image", handler.UploadProductImage)
			mAuth.PUT("/password", handler.MerchantResetWebPasswordByMiniProgram)
			mAuth.PUT("/web/password", handler.MerchantUpdateWebPassword)
			// 店铺
			mAuth.POST("/shop", handler.CreateShop)           // 创建店铺
			mAuth.GET("/shop", handler.GetMerchantShop)       // 查看店铺信息
			mAuth.PUT("/shop", handler.UpdateShop)            // 更新店铺信息
			mAuth.PUT("/shop/status", handler.ToggleShopOpen) // 切换营业状态
			mAuth.POST("/shop/pos-bind-token/rotate", handler.RotateShopPOSBindToken)
			mAuth.PUT("/shop/auto-accept", handler.ToggleShopAutoAccept)
			mAuth.GET("/shop/tables", handler.ListMerchantShopTables)
			mAuth.POST("/shop/table", handler.CreateMerchantShopTable)
			mAuth.PUT("/shop/table/:id", handler.UpdateMerchantShopTable)
			mAuth.DELETE("/shop/table/:id", handler.DeleteMerchantShopTable)

			// 菜单分类
			mAuth.POST("/category", handler.CreateCategory)
			mAuth.PUT("/category/:id", handler.UpdateCategory)
			mAuth.DELETE("/category/:id", handler.DeleteCategory)

			// 菜品
			mAuth.POST("/product", handler.CreateProduct)
			mAuth.PUT("/product/:id", handler.UpdateProduct)
			mAuth.DELETE("/product/:id", handler.DeleteProduct)

			// 订单
			mAuth.GET("/orders", handler.GetMerchantOrders)
			mAuth.GET("/order/:id", handler.GetMerchantOrderDetail)
			mAuth.PUT("/order/:id/accept", handler.AcceptOrder)
			mAuth.PUT("/order/:id/complete", handler.CompleteOrder)
		}
	}

	// --------------- POS 端 ---------------
	posGroup := api.Group("/pos")
	{
		posGroup.POST("/login", handler.POSLogin) // POS 登录
		posGroup.POST("/menu", handler.POSGetMenu)
		posGroup.POST("/order", handler.POSCreateOrder)
		posGroup.GET("/order/:id", handler.POSGetOrderDetail)
		posGroup.POST("/order/:id/pay", handler.POSCreateOrderPay)
		posGroup.POST("/order/:id/cancel", handler.POSCancelOrder)
		posGroup.GET("/ws", handler.POSWebSocket) // WebSocket 连接
	}

	// --------------- 后台管理 ---------------
	adminGroup := api.Group("/admin")
	{
		adminGroup.POST("/login", handler.AdminLogin) // 管理员登录

		aAuth := adminGroup.Group("", middleware.JWTAuth(), middleware.RoleAuth("admin"))
		{
			aAuth.GET("/system-configs", handler.AdminGetSystemConfigs)
			aAuth.PUT("/system-config", handler.AdminUpdateSystemConfig)
			aAuth.GET("/recharge-orders", handler.AdminGetRechargeOrders)
			aAuth.GET("/recharge-order/:id", handler.AdminGetRechargeOrderDetail)
			aAuth.POST("/recharge-order/:id/cleanup-remaining", handler.AdminCleanupRechargeOrderRemaining)
			aAuth.GET("/recharge-refund-requests", handler.AdminGetRechargeRefundRequests)
			aAuth.GET("/recharge-refund-request/:id", handler.AdminGetRechargeRefundRequestDetail)
			aAuth.POST("/recharge-refund-request", handler.AdminCreateRechargeRefundRequest)
			aAuth.PUT("/recharge-refund-request/:id/review", handler.AdminReviewRechargeRefundRequest)
			aAuth.PUT("/recharge-refund-request/:id/sync-status", handler.AdminSyncRechargeRefundRequestStatus)

			aAuth.GET("/recharge-activities", handler.AdminGetRechargeActivities)
			aAuth.POST("/recharge-activity", handler.AdminCreateRechargeActivity)
			aAuth.PUT("/recharge-activity/:id", handler.AdminUpdateRechargeActivity)
			aAuth.DELETE("/recharge-activity/:id", handler.AdminDeleteRechargeActivity)

			// 商家审核
			aAuth.GET("/merchants", handler.AdminGetMerchants)
			aAuth.GET("/merchant/:id", handler.AdminGetMerchantDetail)
			aAuth.DELETE("/merchant/:id", handler.AdminDeleteMerchant)
			aAuth.GET("/users", handler.AdminGetUsers)
			aAuth.GET("/user/:id", handler.AdminGetUserDetail)
			aAuth.POST("/user/:id/cleanup-historical-balance", handler.AdminCleanupUserHistoricalBalance)
			aAuth.GET("/shop/:id", handler.AdminGetShopDetail)
			aAuth.PUT("/shop/:id/audit", handler.AdminAuditShop) // 审核店铺
			aAuth.GET("/shop/:id/audit-records", handler.AdminGetShopAuditRecords)
			aAuth.DELETE("/shop/:id", handler.AdminDeleteShop)

			// 店铺管理
			aAuth.GET("/shops", handler.AdminGetShops)

			// 订单查看
			aAuth.GET("/orders", handler.AdminGetOrders)
			aAuth.GET("/order/:id", handler.AdminGetOrderDetail)

			// 数据概览
			aAuth.GET("/dashboard", handler.AdminDashboard)
		}
	}

	return r
}
