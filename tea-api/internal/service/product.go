package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/rabbitmq"
)

type ProductService struct {
	db *gorm.DB
}

var ErrProductNotFound = errors.New("product not found")

func NewProductService() *ProductService {
	return &ProductService{
		db: database.GetDB(),
	}
}

// CreateCategory 创建商品分类
func (s *ProductService) CreateCategory(category *model.Category) error {
	if err := s.db.Create(category).Error; err != nil {
		return fmt.Errorf("创建分类失败: %w", err)
	}
	return nil
}

// GetCategories 获取商品分类列表
func (s *ProductService) GetCategories(parentID *uint, status *int) ([]*model.Category, error) {
	var categories []*model.Category
	query := s.db.Model(&model.Category{})

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("sort ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("获取分类列表失败: %w", err)
	}

	return categories, nil
}

// UpdateCategory 更新商品分类
func (s *ProductService) UpdateCategory(id uint, updates map[string]interface{}) error {
	if err := s.db.Model(&model.Category{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新分类失败: %w", err)
	}
	return nil
}

// DeleteCategory 删除商品分类
func (s *ProductService) DeleteCategory(id uint) error {
	// 检查是否有子分类
	var count int64
	if err := s.db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("检查子分类失败: %w", err)
	}

	if count > 0 {
		return errors.New("该分类下还有子分类，无法删除")
	}

	// 检查是否有商品
	if err := s.db.Model(&model.Product{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("检查商品失败: %w", err)
	}

	if count > 0 {
		return errors.New("该分类下还有商品，无法删除")
	}

	if err := s.db.Delete(&model.Category{}, id).Error; err != nil {
		return fmt.Errorf("删除分类失败: %w", err)
	}

	return nil
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(product *model.Product) error {
	if err := s.db.Create(product).Error; err != nil {
		return fmt.Errorf("创建商品失败: %w", err)
	}

	// 发布商品创建消息到RabbitMQ
	go func() {
		msg := rabbitmq.OrderMessage{
			OrderID:   product.ID,
			UserID:    0, // 系统操作
			Action:    "product_created",
			Status:    "active",
			Timestamp: time.Now().Unix(),
		}

		if err := rabbitmq.PublishOrderMessage(msg); err != nil {
			fmt.Printf("发布商品创建消息失败: %v\n", err)
		}
	}()

	return nil
}

// GetProducts 获取商品列表
func (s *ProductService) GetProducts(page, limit int, categoryID *uint, status, keyword string, brandID *uint) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64

	query := s.db.Model(&model.Product{}).Preload("Category").Preload("Brand")

	// 条件过滤
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if brandID != nil {
		query = query.Where("brand_id = ?", *brandID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取商品总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("获取商品列表失败: %w", err)
	}

	return products, total, nil
}

// ProductWithStore 包含门店维度字段的商品信息
type ProductWithStore struct {
	model.Product
	StoreStock         *int    `json:"store_stock"`
	StorePriceOverride *string `json:"store_price_override"`
	BrandName          *string `json:"brand_name"`
}

// GetProductsForStore 获取指定门店维度的商品列表（含门店库存与覆盖价）
func (s *ProductService) GetProductsForStore(page, limit int, categoryID *uint, status, keyword string, brandID *uint, storeID uint) ([]*ProductWithStore, int64, error) {
	var list []*ProductWithStore
	var total int64

	if storeID == 0 {
		return nil, 0, errors.New("storeID 必须大于0")
	}

	// 基础查询（计算总数）
	base := s.db.Table("products p")
	if categoryID != nil {
		base = base.Where("p.category_id = ?", *categoryID)
	}
	if status != "" {
		base = base.Where("p.status = ?", status)
	}
	if keyword != "" {
		base = base.Where("p.name LIKE ? OR p.description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if brandID != nil {
		base = base.Where("p.brand_id = ?", *brandID)
	}
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取商品总数失败: %w", err)
	}

	// 具体查询（左连接门店商品）
	query := s.db.Table("products p").
		Select("p.*, sp.stock AS store_stock, sp.price_override AS store_price_override, b.name AS brand_name").
		Joins("LEFT JOIN store_products sp ON sp.product_id = p.id AND sp.store_id = ?", storeID).
		Joins("LEFT JOIN brands b ON b.id = p.brand_id")

	if categoryID != nil {
		query = query.Where("p.category_id = ?", *categoryID)
	}
	if status != "" {
		query = query.Where("p.status = ?", status)
	}
	if keyword != "" {
		query = query.Where("p.name LIKE ? OR p.description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if brandID != nil {
		query = query.Where("p.brand_id = ?", *brandID)
	}

	offset := (page - 1) * limit
	if err := query.Order("p.id DESC").Offset(offset).Limit(limit).Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("获取门店商品列表失败: %w", err)
	}
	return list, total, nil
}

// GetProduct 获取商品详情
func (s *ProductService) GetProduct(id uint) (*model.Product, error) {
	var product model.Product
	if err := s.db.Preload("Category").Preload("Brand").Preload("Skus").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("获取商品详情失败: %w", err)
	}

	return &product, nil
}

// ProductWithStoreDetail 商品详情（含门店维度字段）
type ProductWithStoreDetail struct {
	model.Product
	StoreStock         *int    `json:"store_stock"`
	StorePriceOverride *string `json:"store_price_override"`
}

// GetProductForStore 获取指定门店维度的商品详情
func (s *ProductService) GetProductForStore(id, storeID uint) (*ProductWithStoreDetail, error) {
	p, err := s.GetProduct(id)
	if err != nil {
		return nil, err
	}
	detail := &ProductWithStoreDetail{Product: *p}
	// 查询门店商品绑定
	var sp model.StoreProduct
	if err := s.db.Where("store_id = ? AND product_id = ?", storeID, id).First(&sp).Error; err == nil {
		stock := sp.Stock
		detail.StoreStock = &stock
		// 仅在覆盖价>0时返回字符串
		if sp.PriceOverride.GreaterThan(decimalZero()) {
			s := sp.PriceOverride.String()
			detail.StorePriceOverride = &s
		}
	}
	return detail, nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(id uint, updates map[string]interface{}) error {
	if err := s.db.Model(&model.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新商品失败: %w", err)
	}

	return nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(id uint) error {
	// 检查是否有未完成的订单
	var count int64
	if err := s.db.Model(&model.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("order_items.product_id = ? AND orders.status NOT IN ('completed', 'cancelled')", id).
		Count(&count).Error; err != nil {
		return fmt.Errorf("检查订单失败: %w", err)
	}

	if count > 0 {
		return errors.New("该商品有未完成的订单，无法删除")
	}

	if err := s.db.Delete(&model.Product{}, id).Error; err != nil {
		return fmt.Errorf("删除商品失败: %w", err)
	}

	return nil
}

// UpdateProductStock 更新商品库存
func (s *ProductService) UpdateProductStock(id uint, stock int, action string) error {
	var product model.Product
	if err := s.db.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return fmt.Errorf("获取商品失败: %w", err)
	}

	var newStock int
	switch action {
	case "add":
		newStock = product.Stock + stock
	case "sub":
		newStock = product.Stock - stock
		if newStock < 0 {
			return errors.New("库存不足")
		}
	case "set":
		newStock = stock
	default:
		return errors.New("无效的操作类型")
	}

	if err := s.db.Model(&product).Update("stock", newStock).Error; err != nil {
		return fmt.Errorf("更新库存失败: %w", err)
	}

	return nil
}
