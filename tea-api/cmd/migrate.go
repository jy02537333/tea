//go:build demo
// +build demo

// 数据库迁移演示入口（需 -tags demo）
package main

import (
	"fmt"
	"log"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("configs/config.yaml"); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库连接
	database.InitMySQL()

	fmt.Println("开始执行数据库迁移...")

	// 执行自动迁移
	err := database.GetDB().AutoMigrate(
		// 基础模型
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},

		// 商品管理
		&model.Category{},
		&model.Product{},
		&model.ProductSku{},
		&model.ProductImage{},

		// 订单管理
		&model.Store{},
		&model.StoreProduct{},
		&model.Order{},
		&model.OrderItem{},
		&model.Cart{},
		&model.CartItem{},

		// 支付管理
		&model.Payment{},
		&model.Refund{},

		// 外卖平台
		&model.DeliveryOrder{},
		&model.DeliveryPlatformOrder{},

		// 营销活动
		&model.Coupon{},
		&model.UserCoupon{},
		&model.Activity{},
		&model.ActivityProduct{},
		&model.ActivityRegistration{},

		// 配置管理
		&model.SystemConfig{},
		&model.Banner{},

		// 统计分析
		&model.OrderStatistics{},
		&model.ProductStatistics{},
		&model.UserStatistics{},

		// 日志管理
		&model.AccessLog{},
		&model.OperationLog{},

		// 提现管理
		&model.WithdrawRecord{},
		&model.WechatTransferRecord{},
	)

	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("数据库迁移完成!")

	// 可选：创建一些初始数据
	createInitialData()
}

func createInitialData() {
	db := database.GetDB()

	fmt.Println("创建初始数据...")

	// 创建默认角色
	roles := []model.Role{
		{Name: "admin", DisplayName: "管理员", Description: "系统管理员"},
		{Name: "user", DisplayName: "普通用户", Description: "普通用户"},
	}

	for _, role := range roles {
		var existingRole model.Role
		if err := db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			// 角色不存在，创建它
			if err := db.Create(&role).Error; err != nil {
				fmt.Printf("创建角色 %s 失败: %v\n", role.Name, err)
			} else {
				fmt.Printf("创建角色 %s 成功\n", role.Name)
			}
		}
	}

	// 创建默认商品分类
	categories := []model.Category{
		{Name: "茶叶", Description: "各类优质茶叶", Sort: 1},
		{Name: "茶具", Description: "精美茶具用品", Sort: 2},
		{Name: "茶点", Description: "精致茶点小食", Sort: 3},
		{Name: "礼品", Description: "茶礼品套装", Sort: 4},
	}

	for _, category := range categories {
		var existingCategory model.Category
		if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			// 分类不存在，创建它
			if err := db.Create(&category).Error; err != nil {
				fmt.Printf("创建分类 %s 失败: %v\n", category.Name, err)
			} else {
				fmt.Printf("创建分类 %s 成功\n", category.Name)
			}
		}
	}

	fmt.Println("初始数据创建完成!")
}
