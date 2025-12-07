//go:build ignore
// +build ignore

// 该文件为演示/本地工具入口，默认不参与构建，避免与正式 main 冲突。
package main

import (
	"fmt"
	"log"
	"tea-test/pkg/env"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt int64          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64          `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category 商品分类模型
type Category struct {
	BaseModel
	Name        string `gorm:"type:varchar(50);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Status      int    `gorm:"type:tinyint;default:1" json:"status"`
}

// Product 商品模型
type Product struct {
	BaseModel
	CategoryID  uint            `gorm:"index;not null" json:"category_id"`
	Name        string          `gorm:"type:varchar(100);not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int             `gorm:"default:0" json:"stock"`
	Sales       int             `gorm:"default:0" json:"sales"`
	Status      int             `gorm:"type:tinyint;default:1" json:"status"`

	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

var db *gorm.DB

func initDatabase() error {
	// support TEA_DSN or per-value env overrides
	dsn := env.Get("TEA_DSN", "")
	if dsn == "" {
		host := env.Get("TEA_DATABASE_HOST", "127.0.0.1")
		port := env.Get("TEA_DATABASE_PORT", "3308")
		user := env.Get("TEA_DATABASE_USERNAME", "root")
		pass := env.Get("TEA_DATABASE_PASSWORD", "gs963852")
		dbname := env.Get("TEA_DATABASE_DBNAME", "tea_shop")
		charset := env.Get("TEA_DATABASE_CHARSET", "utf8mb4")
		parseTime := env.Get("TEA_DATABASE_PARSETIME", "True")
		loc := env.Get("TEA_DATABASE_LOC", "Local")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s", user, pass, host, port, dbname, charset, parseTime, loc)
	}

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("✅ 数据库连接成功")

	// 自动迁移
	log.Println("🔄 开始数据库迁移...")
	err = db.AutoMigrate(&Category{}, &Product{})
	if err != nil {
		log.Printf("❌ 数据库迁移失败: %v", err)
		return err
	}
	log.Println("✅ 数据库迁移完成")

	// 初始化示例数据
	initSampleData()
	return nil
}

func initSampleData() {
	var count int64
	db.Model(&Category{}).Count(&count)
	if count > 0 {
		log.Println("📊 数据库已有数据，跳过初始化")
		return
	}

	log.Println("🌱 开始初始化示例数据...")

	// 创建分类
	categories := []Category{
		{Name: "绿茶", Description: "清香淡雅的绿茶系列", Status: 1},
		{Name: "红茶", Description: "香醇浓郁的红茶系列", Status: 1},
		{Name: "乌龙茶", Description: "半发酵的乌龙茶系列", Status: 1},
	}

	for _, category := range categories {
		db.Create(&category)
	}

	// 创建产品
	products := []Product{
		{
			CategoryID:  1,
			Name:        "西湖龙井",
			Description: "正宗西湖龙井茶，清香甘甜",
			Price:       decimal.NewFromFloat(168.00),
			Stock:       50,
			Status:      1,
		},
		{
			CategoryID:  1,
			Name:        "碧螺春",
			Description: "江苏苏州洞庭碧螺春，香气浓郁",
			Price:       decimal.NewFromFloat(138.00),
			Stock:       30,
			Status:      1,
		},
		{
			CategoryID:  2,
			Name:        "正山小种",
			Description: "福建武夷山正宗正山小种红茶",
			Price:       decimal.NewFromFloat(128.50),
			Stock:       45,
			Status:      1,
		},
	}

	for _, product := range products {
		db.Create(&product)
	}

	log.Println("✅ 示例数据初始化完成")
}

func main() {
	fmt.Println("🚀 启动茶心阁API服务器 (数据库版本)...")

	// 初始化数据库
	if err := initDatabase(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 设置Gin
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由
	api := r.Group("/api/v1")

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success":  true,
			"message":  "Tea API Server (Database Version) is running",
			"database": "MySQL Connected",
		})
	})

	// 产品管理
	api.GET("/products", getProducts)
	api.GET("/products/:id", getProduct)
	api.POST("/products", createProduct)
	api.PUT("/products/:id", updateProduct)
	api.DELETE("/products/:id", deleteProduct)

	// 分类管理
	api.GET("/categories", getCategories)
	api.GET("/categories/:id", getCategory)
	api.POST("/categories", createCategory)
	api.PUT("/categories/:id", updateCategory)
	api.DELETE("/categories/:id", deleteCategory)

	port := ":9292"
	fmt.Printf("🔗 服务器运行在: http://localhost%s\n", port)
	fmt.Printf("🔗 健康检查: http://localhost%s/api/v1/health\n", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// API处理函数
func getProducts(c *gin.Context) {
	var products []Product
	result := db.Preload("Category").Find(&products)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "message": "获取产品列表失败", "error": result.Error.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": products, "count": len(products)})
}

func getProduct(c *gin.Context) {
	id := c.Param("id")
	var product Product
	result := db.Preload("Category").First(&product, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"success": false, "message": "产品未找到"})
		} else {
			c.JSON(500, gin.H{"success": false, "message": "获取产品失败", "error": result.Error.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"success": true, "data": product})
}

func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "请求数据格式错误", "error": err.Error()})
		return
	}

	result := db.Create(&product)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "message": "产品创建失败", "error": result.Error.Error()})
		return
	}

	db.Preload("Category").First(&product, product.ID)
	c.JSON(201, gin.H{"success": true, "message": "产品创建成功", "data": product})
}

func updateProduct(c *gin.Context) {
	id := c.Param("id")
	var product Product

	if db.First(&product, id).Error != nil {
		c.JSON(404, gin.H{"success": false, "message": "产品未找到"})
		return
	}

	var updateData Product
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "请求数据格式错误", "error": err.Error()})
		return
	}

	if err := db.Model(&product).Updates(updateData).Error; err != nil {
		c.JSON(500, gin.H{"success": false, "message": "产品更新失败", "error": err.Error()})
		return
	}

	db.Preload("Category").First(&product, product.ID)
	c.JSON(200, gin.H{"success": true, "message": "产品更新成功", "data": product})
}

func deleteProduct(c *gin.Context) {
	id := c.Param("id")
	result := db.Delete(&Product{}, id)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "message": "产品删除失败", "error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"success": false, "message": "产品未找到"})
		return
	}
	c.JSON(200, gin.H{"success": true, "message": "产品删除成功"})
}

func getCategories(c *gin.Context) {
	var categories []Category
	if err := db.Find(&categories).Error; err != nil {
		c.JSON(500, gin.H{"success": false, "message": "获取分类列表失败", "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": categories, "count": len(categories)})
}

func getCategory(c *gin.Context) {
	id := c.Param("id")
	var category Category
	result := db.First(&category, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"success": false, "message": "分类未找到"})
		} else {
			c.JSON(500, gin.H{"success": false, "message": "获取分类失败", "error": result.Error.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"success": true, "data": category})
}

func createCategory(c *gin.Context) {
	var category Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "请求数据格式错误", "error": err.Error()})
		return
	}

	if err := db.Create(&category).Error; err != nil {
		c.JSON(500, gin.H{"success": false, "message": "分类创建失败", "error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"success": true, "message": "分类创建成功", "data": category})
}

func updateCategory(c *gin.Context) {
	id := c.Param("id")
	var category Category

	if db.First(&category, id).Error != nil {
		c.JSON(404, gin.H{"success": false, "message": "分类未找到"})
		return
	}

	var updateData Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "请求数据格式错误", "error": err.Error()})
		return
	}

	if err := db.Model(&category).Updates(updateData).Error; err != nil {
		c.JSON(500, gin.H{"success": false, "message": "分类更新失败", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "分类更新成功", "data": category})
}

func deleteCategory(c *gin.Context) {
	id := c.Param("id")

	// 检查分类下是否有产品
	var productCount int64
	db.Model(&Product{}).Where("category_id = ?", id).Count(&productCount)
	if productCount > 0 {
		c.JSON(400, gin.H{"success": false, "message": "该分类下还有产品，无法删除"})
		return
	}

	result := db.Delete(&Category{}, id)
	if result.Error != nil {
		c.JSON(500, gin.H{"success": false, "message": "分类删除失败", "error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"success": false, "message": "分类未找到"})
		return
	}
	c.JSON(200, gin.H{"success": true, "message": "分类删除成功"})
}

// (removed local getEnv) using shared pkg/env Get
