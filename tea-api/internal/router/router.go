package router

import (
	"fmt"
	"math/rand"
	"time"

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

	// API路由组
	api := r.Group("/api/v1")

	// 用户相关路由
	userGroup := api.Group("/user")
	{
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/dev-login", userHandler.DevLogin)
		userGroup.POST("/password", middleware.AuthMiddleware(), userHandler.ChangePassword)
		userGroup.POST("/refresh", userHandler.Refresh)
		userGroup.GET("/interest-records", middleware.AuthMiddleware(), accrualHandler.UserInterestRecords)
		userGroup.GET("/info", middleware.AuthMiddleware(), userHandler.GetUserInfo)
		userGroup.PUT("/info", middleware.AuthMiddleware(), userHandler.UpdateUserInfo)
		userGroup.GET("/:id", userHandler.GetUserByID)
	}

	// 管理员路由（仅管理员可访问）
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireRoles("admin"))
	{
		adminGroup.GET("/users", userHandler.AdminListUsers)
		// 门店订单统计
		adminGroup.GET("/stores/:id/orders/stats", storeHandler.OrderStats)
		// 门店库存管理
		adminGroup.GET("/stores/:id/products", invHandler.List)
		adminGroup.POST("/stores/:id/products", invHandler.Upsert)
		adminGroup.DELETE("/stores/:id/products/:pid", invHandler.Delete)

		// 管理端订单接口（列表 / 导出 / 详情）
		adminGroup.GET("/orders", orderHandler.AdminList)
		adminGroup.GET("/orders/export", orderHandler.AdminExport)
		adminGroup.GET("/orders/:id", orderHandler.AdminDetail)
	}

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
	cartGroup.Use(middleware.AuthMiddleware())
	{
		cartGroup.GET("", cartHandler.List)
		cartGroup.POST("/items", cartHandler.AddItem)
		cartGroup.PUT("/items/:id", cartHandler.UpdateQuantity)
		cartGroup.DELETE("/items/:id", cartHandler.Remove)
		cartGroup.DELETE("/clear", cartHandler.Clear)
	}

	// 订单相关路由（需要登录）
	orderGroup := api.Group("/orders")
	orderGroup.Use(middleware.AuthMiddleware())
	{
		orderGroup.POST("/from-cart", orderHandler.CreateFromCart)
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
		storeGroup.GET("/:id", storeHandler.Get)
		storeGroup.POST("", middleware.AuthMiddleware(), storeHandler.Create)
		storeGroup.PUT("/:id", middleware.AuthMiddleware(), storeHandler.Update)
		storeGroup.DELETE("/:id", middleware.AuthMiddleware(), storeHandler.Delete)
	}

	// 优惠券相关路由
	couponGroup := api.Group("/coupons")
	{
		couponGroup.GET("", couponHandler.ListCoupons)
		couponGroup.POST("", middleware.AuthMiddleware(), couponHandler.CreateCoupon)
		couponGroup.POST("/grant", middleware.AuthMiddleware(), couponHandler.Grant)
	}

	// 用户优惠券（需要登录）
	userCouponGroup := api.Group("/user")
	userCouponGroup.Use(middleware.AuthMiddleware())
	{
		userCouponGroup.GET("/coupons", couponHandler.ListMyCoupons)
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

	// 兼容旧前端路由：提供 /auth/* 别名（登录/开发登录/用户信息）
	// 以及一个简单的 /auth/captcha 开发用接口（返回 id + code），便于 Admin-FE 的开发登录流程使用
	// compatibility: accept legacy auth/login which may send form data
	api.POST("/auth/login", handler.AuthLogin)
	api.POST("/auth/dev-login", userHandler.DevLogin)
	api.GET("/auth/me", middleware.AuthMiddleware(), userHandler.GetUserInfo)
	api.GET("/auth/captcha", func(c *gin.Context) {
		// 简易开发用验证码：返回随机 4 位数字和一个瞬时 id
		rand.Seed(time.Now().UnixNano())
		code := fmt.Sprintf("%04d", rand.Intn(10000))
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		c.JSON(200, gin.H{"id": id, "code": code})
	})

	// NOTE: legacy API-Server compatibility routes were intentionally removed to avoid
	// registering duplicate handlers that are already implemented in the main codebase.
	// If specific dev-only mock routes are required, add them selectively in a
	// dedicated dev-only router to avoid conflicts.

	// 支付（模拟）
	payGroup := api.Group("/payment")
	payGroup.Use(middleware.AuthMiddleware())
	{
		payGroup.POST("/intents", paymentHandler.CreateIntent)
	}
	// 模拟回调（仅开发环境）
	api.POST("/payment/mock-callback", paymentHandler.MockCallback)

	return r
}
