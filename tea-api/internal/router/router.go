package router

import (
	"github.com/gin-gonic/gin"

	"tea-api/internal/handler"
	"tea-api/internal/middleware"
	"tea-api/internal/service"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()

	// 添加中间件：顺序为请求ID -> 恢复 -> 访问日志 -> CORS -> 认证
	r.Use(middleware.RequestIDMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.DetailedAccessLogMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 初始化处理器
	userHandler := handler.NewUserHandler()
	accrualHandler := handler.NewAccrualHandler()
	rbacHandler := handler.NewRBACHandler()
	logsHandler := handler.NewLogsHandler()
	refundHandler := handler.NewRefundHandler()
	financeReportHandler := handler.NewFinanceReportHandler()
	commissionAdminHandler := handler.NewCommissionAdminHandler()
	membershipAdminHandler := handler.NewMembershipAdminHandler()
	membershipHandler := handler.NewMembershipHandler()
	dashboardHandler := handler.NewDashboardHandler()
	productHandler := handler.NewProductHandler(
		service.NewProductService(),
	)
	cartHandler := handler.NewCartHandler()
	orderHandler := handler.NewOrderHandler()
	paymentHandler := handler.NewPaymentHandler()
	paymentAdminHandler := handler.NewPaymentAdminHandler()
	withdrawAdminHandler := handler.NewWithdrawAdminHandler()
	couponHandler := handler.NewCouponHandler()
	storeHandler := handler.NewStoreHandler()
	invHandler := handler.NewStoreInventoryHandler()
	modelHandler := handler.NewModelHandler()
	uploadHandler := handler.NewUploadHandler()
	activityHandler := handler.NewActivityHandler()
	ticketHandler := handler.NewTicketHandler()

	// API路由组
	api := r.Group("/api/v1")

	// 统一鉴权登录入口（Sprint B）
	handler.RegisterAuthRoutes(api)

	// Sprint B: 我的/个人中心聚合（最小连通性）
	api.GET("/users/me/summary", middleware.AuthJWT(), handler.GetUserSummary)

	// Sprint B: 用户钱包（余额与流水）
	api.GET("/wallet", middleware.AuthJWT(), handler.GetMyWallet)
	api.GET("/wallet/transactions", middleware.AuthJWT(), handler.ListMyWalletTransactions)

	// Sprint B: 用户提现账户与申请
	api.GET("/wallet/bank-accounts", middleware.AuthJWT(), handler.ListMyBankAccounts)
	api.POST("/wallet/bank-accounts", middleware.AuthJWT(), handler.CreateMyBankAccount)
	api.DELETE("/wallet/bank-accounts/:id", middleware.AuthJWT(), handler.DeleteMyBankAccount)
	// 两种入口：按 PRD 支持 /wallet/withdrawals 以及 /users/{id}/withdrawals
	api.POST("/wallet/withdrawals", middleware.AuthJWT(), handler.CreateMyWithdrawal)
	api.GET("/users/:id/withdrawals", middleware.AuthJWT(), handler.ListUserWithdrawals)
	api.POST("/users/:id/withdrawals", middleware.AuthJWT(), handler.CreateUserWithdrawal)

	// Sprint B: 积分查询与流水
	api.GET("/points", middleware.AuthJWT(), handler.GetMyPoints)
	api.GET("/points/transactions", middleware.AuthJWT(), handler.ListMyPointsTransactions)

	// Sprint B: 优惠券模板与领取
	api.GET("/coupons/templates", middleware.AuthJWT(), handler.ListCouponTemplates)
	api.POST("/coupons/claim", middleware.AuthJWT(), handler.ClaimCouponFromTemplate)

	// 会员相关（小程序/用户侧只读接口）
	api.GET("/membership-packages", middleware.AuthMiddleware(), membershipHandler.ListPackages)
	api.POST("/membership-orders", middleware.AuthMiddleware(), membershipHandler.CreateOrder)

	// 用户相关路由
	userGroup := api.Group("/user")
	{
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/dev-login", userHandler.DevLogin)
		// 旧路由统一使用 AuthMiddleware（已改为新版密钥校验），避免校验分叉
		userGroup.POST("/password", middleware.AuthMiddleware(), userHandler.ChangePassword)
		userGroup.POST("/refresh", userHandler.Refresh)
		userGroup.GET("/interest-records", middleware.AuthMiddleware(), accrualHandler.UserInterestRecords)
		userGroup.GET("/info", middleware.AuthMiddleware(), userHandler.GetUserInfo)
		userGroup.PUT("/info", middleware.AuthMiddleware(), userHandler.UpdateUserInfo)
		userGroup.GET("/default-address", middleware.AuthMiddleware(), userHandler.GetDefaultAddress)
		userGroup.PUT("/default-address", middleware.AuthMiddleware(), userHandler.UpdateDefaultAddress)
		userGroup.GET("/:id", userHandler.GetUserByID)
		// 当前用户的可用优惠券列表（小程序「我的-优惠券」只读 Tab 使用）
		userGroup.GET("/coupons", middleware.AuthMiddleware(), couponHandler.ListMyCoupons)
	}

	// 管理员路由（仅管理员可访问）
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireRoles("admin"))
	{
		// 后台首页待办统计
		adminGroup.GET("/dashboard/todos", dashboardHandler.Todos)
		// 后台首页摘要（今日）
		adminGroup.GET("/dashboard/summary", dashboardHandler.Summary)
		// 后台首页订单趋势（按天）
		adminGroup.GET("/dashboard/order-trends", dashboardHandler.OrderTrends)
		// 兼容别名（下划线/无连字符），避免部分客户端路径解析异常
		adminGroup.GET("/dashboard/order_trends", dashboardHandler.OrderTrends)
		adminGroup.GET("/dashboard/ordertrends", dashboardHandler.OrderTrends)
		// 探针：验证 dashboard 前缀可达
		adminGroup.GET("/dashboard/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"data": gin.H{"ok": true}})
		})

		adminGroup.GET("/users", userHandler.AdminListUsers)
		adminGroup.POST("/users", userHandler.AdminCreateUser)
		adminGroup.PUT("/users/:id", userHandler.AdminUpdateUser)
		adminGroup.POST("/users/:id/reset-password", userHandler.AdminResetPassword)
		adminGroup.POST("/uploads", uploadHandler.UploadMedia)
		// 门店订单统计
		adminGroup.GET("/stores/:id/orders/stats", storeHandler.OrderStats)
		// 门店订单列表（按门店维度查看订单）
		adminGroup.GET("/stores/:id/orders", orderHandler.AdminStoreOrders)
		// 门店库存管理
		adminGroup.GET("/stores/:id/products", invHandler.List)
		adminGroup.POST("/stores/:id/products", invHandler.Upsert)
		adminGroup.DELETE("/stores/:id/products/:pid", invHandler.Delete)

		// 客服工单管理
		adminGroup.GET("/tickets", ticketHandler.List)
		adminGroup.GET("/tickets/:id", ticketHandler.Get)
		adminGroup.POST("/tickets", ticketHandler.Create)
		adminGroup.PUT("/tickets/:id", ticketHandler.Update)

		// 管理端订单接口（列表 / 导出 / 详情）
		adminGroup.GET("/orders", orderHandler.AdminList)
		adminGroup.GET("/orders/export", orderHandler.AdminExport)
		adminGroup.GET("/orders/:id", orderHandler.AdminDetail)
		// 门店后台：接受/拒绝订单
		adminGroup.POST("/orders/:id/accept", middleware.RequirePermission("order:accept"), orderHandler.AdminAccept)
		adminGroup.POST("/orders/:id/reject", middleware.RequirePermission("order:reject"), orderHandler.AdminReject)

		// 会员与合伙人配置
		adminGroup.GET("/membership-packages", membershipAdminHandler.ListPackages)
		adminGroup.POST("/membership-packages", membershipAdminHandler.CreatePackage)
		adminGroup.PUT("/membership-packages/:id", membershipAdminHandler.UpdatePackage)
		adminGroup.DELETE("/membership-packages/:id", membershipAdminHandler.DeletePackage)

		adminGroup.GET("/partner-levels", membershipAdminHandler.ListPartnerLevels)
		adminGroup.POST("/partner-levels", membershipAdminHandler.CreatePartnerLevel)
		adminGroup.PUT("/partner-levels/:id", membershipAdminHandler.UpdatePartnerLevel)
		adminGroup.DELETE("/partner-levels/:id", membershipAdminHandler.DeletePartnerLevel)

		// 提现申请审批（管理员）别名路径，复用 withdraws 处理器
		adminGroup.GET("/withdrawals", withdrawAdminHandler.List)
		adminGroup.POST("/withdrawals/:id/approve", withdrawAdminHandler.Approve)
		adminGroup.POST("/withdrawals/:id/reject", withdrawAdminHandler.Reject)
	}

	// 调试与容错：为订单趋势提供一个仅鉴权、不做角色校验的别名，便于前端联调
	// 如生产环境不需要，可移除该别名。
	api.GET("/admin/dashboard/order-trends-public", middleware.AuthMiddleware(), dashboardHandler.OrderTrends)

	// NoRoute 调试输出：统一打印 404 请求的信息，便于定位未匹配路由
	r.NoRoute(func(c *gin.Context) {
		reqID := c.Writer.Header().Get("X-Request-Id")
		if reqID == "" {
			if v, ok := c.Get("request_id"); ok {
				reqID, _ = v.(string)
			}
		}
		c.JSON(404, gin.H{
			"error":      "not_found",
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"request_id": reqID,
		})
	})

	// 计息相关路由（基于权限控制，允许非 admin 但拥有授权的角色访问）
	accrualGroup := api.Group("/admin")
	accrualGroup.Use(middleware.AuthMiddleware())
	accrualGroup.Use(middleware.OperationLogMiddleware())
	{
		accrualGroup.POST("/accrual/run", middleware.RequirePermission("accrual:run"), accrualHandler.AdminAccrualRun)
		accrualGroup.GET("/accrual/export", middleware.RequirePermission("accrual:export"), accrualHandler.AdminAccrualExport)
		accrualGroup.GET("/accrual/summary", middleware.RequirePermission("accrual:summary"), accrualHandler.AdminAccrualSummary)
	}

	// RBAC 管理接口（只读 + 变更；按权限控制）
	rbacGroup := api.Group("/admin/rbac")
	rbacGroup.Use(middleware.AuthMiddleware())
	rbacGroup.Use(middleware.OperationLogMiddleware())
	{
		rbacGroup.GET("/roles", middleware.RequirePermission("rbac:view"), rbacHandler.ListRoles)
		rbacGroup.GET("/permissions", middleware.RequirePermission("rbac:view"), rbacHandler.ListPermissions)
		rbacGroup.GET("/role-permissions", middleware.RequirePermission("rbac:view"), rbacHandler.ListRolePermissions)
		rbacGroup.GET("/user-permissions", middleware.RequirePermission("rbac:view"), rbacHandler.ListUserPermissions)
		rbacGroup.POST("/cache/invalidate", middleware.RequirePermission("rbac:manage"), rbacHandler.InvalidateCache)

		// RBAC 变更接口（需要 rbac:manage）
		rbacGroup.POST("/role", middleware.RequirePermission("rbac:manage"), rbacHandler.CreateRole)
		rbacGroup.DELETE("/role/:id", middleware.RequirePermission("rbac:manage"), rbacHandler.DeleteRole)
		rbacGroup.POST("/permission", middleware.RequirePermission("rbac:manage"), rbacHandler.CreatePermission)
		rbacGroup.POST("/role/assign-permission", middleware.RequirePermission("rbac:manage"), rbacHandler.AssignPermissionToRole)
		rbacGroup.POST("/role/revoke-permission", middleware.RequirePermission("rbac:manage"), rbacHandler.RevokePermissionFromRole)
		rbacGroup.POST("/role/assign-permissions", middleware.RequirePermission("rbac:manage"), rbacHandler.AssignPermissionsToRole)
		rbacGroup.POST("/user/assign-role", middleware.RequirePermission("rbac:manage"), rbacHandler.AssignRoleToUser)
		rbacGroup.POST("/user/revoke-role", middleware.RequirePermission("rbac:manage"), rbacHandler.RevokeRoleFromUser)
	}

	// 操作日志接口（只读，rbac:view）
	logsGroup := api.Group("/admin/logs")
	logsGroup.Use(middleware.AuthMiddleware())
	{
		logsGroup.GET("/operations", middleware.RequirePermission("rbac:view"), logsHandler.ListOperationLogs)
		logsGroup.GET("/operations/export", middleware.RequirePermission("rbac:view"), logsHandler.ExportOperationLogs)
		logsGroup.GET("/access", middleware.RequirePermission("rbac:view"), logsHandler.ListAccessLogs)
		logsGroup.GET("/access/export", middleware.RequirePermission("rbac:view"), logsHandler.ExportAccessLogs)
	}

	// 退款记录（只读列表与导出，按退款权限控制）
	refundsGroup := api.Group("/admin/refunds")
	refundsGroup.Use(middleware.AuthMiddleware())
	{
		refundsGroup.GET("", middleware.RequirePermission("order:refund"), refundHandler.ListRefunds)
		refundsGroup.GET("/export", middleware.RequirePermission("order:refund"), refundHandler.ExportRefunds)
	}

	// 支付记录（财务流水，只读列表与导出，与退款同权限控制）
	paymentsGroup := api.Group("/admin/payments")
	paymentsGroup.Use(middleware.AuthMiddleware())
	{
		paymentsGroup.GET("", middleware.RequirePermission("order:refund"), paymentAdminHandler.ListPayments)
		paymentsGroup.GET("/export", middleware.RequirePermission("order:refund"), paymentAdminHandler.ExportPayments)
	}

	// 提现记录（财务）
	withdrawGroup := api.Group("/admin/withdraws")
	withdrawGroup.Use(middleware.AuthMiddleware())
	{
		withdrawGroup.GET("", middleware.RequirePermission("order:refund"), withdrawAdminHandler.List)
		withdrawGroup.GET("/export", middleware.RequirePermission("order:refund"), withdrawAdminHandler.Export)
		withdrawGroup.POST("/:id/approve", middleware.RequirePermission("order:refund"), withdrawAdminHandler.Approve)
		withdrawGroup.POST("/:id/complete", middleware.RequirePermission("order:refund"), withdrawAdminHandler.Complete)
		withdrawGroup.POST("/:id/reject", middleware.RequirePermission("order:refund"), withdrawAdminHandler.Reject)
	}

	// 财务报表（对账概要/导出）
	financeGroup := api.Group("/admin/finance")
	financeGroup.Use(middleware.AuthMiddleware())
	{
		financeGroup.GET("/summary", middleware.RequirePermission("order:refund"), financeReportHandler.Summary)
		financeGroup.GET("/summary/export", middleware.RequirePermission("order:refund"), financeReportHandler.ExportSummary)
		// 支付对账差异
		financeGroup.GET("/reconcile/diff", middleware.RequirePermission("order:refund"), financeReportHandler.ReconcileDiff)
		financeGroup.GET("/reconcile/diff/export", middleware.RequirePermission("order:refund"), financeReportHandler.ExportReconcileDiff)
		// 佣金解冻手动触发，仅限具备财务权限的账号
		financeGroup.POST("/commission/release", middleware.RequirePermission("order:refund"), commissionAdminHandler.TriggerRelease)
		// 按订单一键回滚未提现佣金
		financeGroup.POST("/commission/reverse-order", middleware.RequirePermission("order:refund"), commissionAdminHandler.ReverseOrder)
	}

	// 商品分类路由
	categoryGroup := api.Group("/categories")
	{
		categoryGroup.POST("", middleware.AuthMiddleware(), productHandler.CreateCategory)
		categoryGroup.GET("", productHandler.GetCategories)
		categoryGroup.PUT("/:id", middleware.AuthMiddleware(), productHandler.UpdateCategory)
		categoryGroup.DELETE("/:id", middleware.AuthMiddleware(), productHandler.DeleteCategory)
	}

	// 商品相关路由
	productGroup := api.Group("/products")
	{
		productGroup.POST("", middleware.AuthMiddleware(), productHandler.CreateProduct)
		productGroup.GET("", productHandler.GetProducts)
		productGroup.GET("/:id", productHandler.GetProduct)
		productGroup.PUT("/:id", middleware.AuthMiddleware(), productHandler.UpdateProduct)
		productGroup.DELETE("/:id", middleware.AuthMiddleware(), productHandler.DeleteProduct)
		productGroup.PUT("/:id/stock", middleware.AuthMiddleware(), productHandler.UpdateProductStock)
	}

	// 购物车相关路由（需要登录）
	cartGroup := api.Group("/cart")
	cartGroup.Use(middleware.AuthJWT())
	{
		cartGroup.GET("", cartHandler.List)
		// 兼容两种写法：POST /cart 与 POST /cart/items 均视为“加入购物车”
		cartGroup.POST("/", cartHandler.AddItem)
		cartGroup.POST("/items", cartHandler.AddItem)
		cartGroup.PUT("/items/:id", cartHandler.UpdateQuantity)
		cartGroup.DELETE("/items/:id", cartHandler.Remove)
		cartGroup.DELETE("/clear", cartHandler.Clear)
	}

	// 订单相关路由（需要登录）
	orderGroup := api.Group("/orders")
	orderGroup.Use(middleware.AuthJWT())
	{
		orderGroup.POST("/from-cart", orderHandler.CreateFromCart)
		orderGroup.POST("/available-coupons", orderHandler.AvailableCoupons)
		orderGroup.GET("", orderHandler.List)
		orderGroup.GET("/:id", orderHandler.Detail)
		orderGroup.POST("/:id/cancel", orderHandler.Cancel)
		orderGroup.POST("/:id/pay", orderHandler.Pay)
		orderGroup.POST("/:id/receive", orderHandler.Receive)
		// 下列操作仅允许具备相应权限（或admin）
		orderGroup.POST("/:id/deliver", middleware.RequirePermission("order:deliver"), orderHandler.Deliver)
		orderGroup.POST("/:id/complete", middleware.RequirePermission("order:complete"), orderHandler.Complete)
		orderGroup.POST("/:id/admin-cancel", middleware.RequirePermission("order:cancel"), orderHandler.AdminCancel)
		orderGroup.POST("/:id/refund", middleware.RequirePermission("order:refund"), orderHandler.AdminRefund)
		orderGroup.POST("/:id/refund/start", middleware.RequirePermission("order:refund"), orderHandler.AdminRefundStart)
		orderGroup.POST("/:id/refund/confirm", middleware.RequirePermission("order:refund"), orderHandler.AdminRefundConfirm)
	}

	// 门店相关路由
	storeGroup := api.Group("/stores")
	{
		storeGroup.GET("", storeHandler.List)
		storeGroup.GET(":id", storeHandler.Get)
		storeGroup.POST("", middleware.AuthJWT(), storeHandler.Create)
		storeGroup.PUT(":id", middleware.AuthJWT(), storeHandler.Update)
		storeGroup.DELETE(":id", middleware.AuthJWT(), storeHandler.Delete)
		// 门店收款账户管理（需要登录，可按角色控制）
		storeGroup.GET(":id/accounts", middleware.AuthJWT(), middleware.RequirePermission("store:accounts:view"), storeHandler.ListAccounts)
		storeGroup.POST(":id/accounts", middleware.AuthJWT(), middleware.RequirePermission("store:accounts:manage"), storeHandler.CreateAccount)
		storeGroup.PUT(":id/accounts/:accountId", middleware.AuthJWT(), middleware.RequirePermission("store:accounts:manage"), storeHandler.UpdateAccount)
		storeGroup.DELETE(":id/accounts/:accountId", middleware.AuthJWT(), middleware.RequirePermission("store:accounts:manage"), storeHandler.DeleteAccount)
		// 门店钱包与提现接口（需要登录，后续可按角色细化权限）
		storeGroup.GET(":id/wallet", middleware.AuthJWT(), middleware.RequirePermission("store:wallet:view"), storeHandler.Wallet)
		storeGroup.GET(":id/withdraws", middleware.AuthJWT(), middleware.RequirePermission("store:withdraw:view"), storeHandler.ListWithdraws)
		storeGroup.POST(":id/withdraws", middleware.AuthJWT(), middleware.RequirePermission("store:withdraw:apply"), storeHandler.ApplyWithdraw)
		// 门店优惠券接口（需要登录，后续可按角色细化权限）
		storeGroup.GET(":id/coupons", middleware.AuthJWT(), middleware.RequirePermission("store:coupons:view"), couponHandler.ListStoreCoupons)
		storeGroup.POST(":id/coupons", middleware.AuthJWT(), middleware.RequirePermission("store:coupons:manage"), couponHandler.CreateStoreCoupon)
		storeGroup.PUT(":id/coupons/:couponId", middleware.AuthJWT(), middleware.RequirePermission("store:coupons:manage"), couponHandler.UpdateStoreCoupon)
		// 门店活动接口（需要登录，后续可按角色细化权限）
		storeGroup.GET(":id/activities", middleware.AuthMiddleware(), activityHandler.ListStoreActivities)
		storeGroup.POST(":id/activities", middleware.AuthMiddleware(), activityHandler.CreateStoreActivity)
		storeGroup.PUT(":id/activities/:activityId", middleware.AuthMiddleware(), activityHandler.UpdateStoreActivity)
		// 门店活动报名接口（需要登录，后续可按角色细化权限）
		storeGroup.GET(":id/activities/:activityId/registrations", middleware.AuthMiddleware(), activityHandler.ListActivityRegistrations)
		storeGroup.POST(":id/activities/:activityId/registrations/:registrationId/refund", middleware.AuthMiddleware(), activityHandler.RefundActivityRegistration)
	}

	// 优惠券相关路由
	couponGroup := api.Group("/coupons")
	{
		couponGroup.GET("", couponHandler.ListCoupons)
		couponGroup.POST("", middleware.AuthMiddleware(), couponHandler.CreateCoupon)
		couponGroup.POST("/grant", middleware.AuthMiddleware(), couponHandler.Grant)
	}

	// 活动（用户侧）
	activityGroup := api.Group("/activities")
	{
		activityGroup.GET("", activityHandler.ListActivities)
		activityGroup.POST(":id/register", middleware.AuthJWT(), middleware.RequirePermission("activity:register"), activityHandler.RegisterActivity)
		activityGroup.POST(":id/register-with-order", middleware.AuthJWT(), middleware.RequirePermission("activity:register"), activityHandler.RegisterActivityWithOrder)
	}

	// 健康检查
	// 模型生成（受 ai.enabled 或 MODEL_API_KEY 控制），带简单限流中间件
	api.POST("/model/generate", middleware.ModelRateLimit(), modelHandler.Generate)

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "茶心阁小程序API服务正常运行",
		})
	})

	// 已迁移至统一的 /api/v1/auth/* 新入口，避免重复与歧义。

	// NOTE: legacy API-Server compatibility routes were intentionally removed to avoid
	// registering duplicate handlers that are already implemented in the main codebase.
	// If specific dev-only mock routes are required, add them selectively in a
	// dedicated dev-only router to avoid conflicts.

	// 支付（模拟）
	payGroup := api.Group("/payment")
	payGroup.Use(middleware.AuthJWT())
	{
		payGroup.POST("/intents", paymentHandler.CreateIntent)
	}

	userPaymentsGroup := api.Group("/payments")
	userPaymentsGroup.Use(middleware.AuthJWT())
	{
		userPaymentsGroup.POST("/unified-order", paymentHandler.UnifiedOrder)
	}
	api.POST("/payments/callback", paymentHandler.Callback)
	// 模拟回调（仅开发环境）
	api.POST("/payment/mock-callback", paymentHandler.MockCallback)

	return r
}
