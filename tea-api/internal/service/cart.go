package service

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type CartService struct {
	db *gorm.DB
}

func NewCartService() *CartService {
	return &CartService{db: database.GetDB()}
}

// IsExclusiveProductForStore checks if a product is bound to the store as an exclusive item.
// biz_type=3 表示门店特供/其他。
func (s *CartService) IsExclusiveProductForStore(storeID uint, productID uint) (bool, error) {
	if storeID == 0 || productID == 0 {
		return false, nil
	}
	const bizTypeExclusive = 3
	var cnt int64
	if err := s.db.Table("store_products").
		Where("store_id = ? AND biz_type = ? AND product_id = ?", storeID, bizTypeExclusive, productID).
		Count(&cnt).Error; err != nil {
		return false, fmt.Errorf("校验门店特供商品失败: %w", err)
	}
	return cnt > 0, nil
}

// FindFirstNonExclusiveProductInCart returns the first product_id in user's cart
// that is NOT bound to the store as an exclusive item.
// If none found, returns 0, nil.
func (s *CartService) FindFirstNonExclusiveProductInCart(storeID uint, userID uint) (uint, error) {
	if storeID == 0 || userID == 0 {
		return 0, nil
	}

	items, err := s.ListItems(userID)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}

	ids := make([]uint, 0, len(items))
	seen := make(map[uint]struct{}, len(items))
	for _, it := range items {
		if it.ProductID == 0 {
			continue
		}
		if _, ok := seen[it.ProductID]; ok {
			continue
		}
		seen[it.ProductID] = struct{}{}
		ids = append(ids, it.ProductID)
	}
	if len(ids) == 0 {
		return 0, nil
	}

	const bizTypeExclusive = 3
	var rows []struct {
		ProductID uint `gorm:"column:product_id"`
	}
	if err := s.db.Table("store_products").
		Select("product_id").
		Where("store_id = ? AND biz_type = ? AND product_id IN ?", storeID, bizTypeExclusive, ids).
		Find(&rows).Error; err != nil {
		return 0, fmt.Errorf("校验购物车商品归属失败: %w", err)
	}

	allowed := make(map[uint]struct{}, len(rows))
	for _, r := range rows {
		if r.ProductID == 0 {
			continue
		}
		allowed[r.ProductID] = struct{}{}
	}

	for _, pid := range ids {
		if _, ok := allowed[pid]; !ok {
			return pid, nil
		}
	}
	return 0, nil
}

// GetOrCreateCart 获取或创建用户购物车
func (s *CartService) GetOrCreateCart(userID uint) (*model.Cart, error) {
	var cart model.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = model.Cart{UserID: userID}
			if err := s.db.Create(&cart).Error; err != nil {
				return nil, fmt.Errorf("创建购物车失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("获取购物车失败: %w", err)
		}
	}
	return &cart, nil
}

// AddItem 向购物车添加商品（同款同SKU合并数量）
func (s *CartService) AddItem(userID uint, productID uint, skuID *uint, quantity int) (*model.CartItem, error) {
	if quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}

	// 校验商品
	var product model.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		return nil, fmt.Errorf("获取商品失败: %w", err)
	}
	if product.Status != 1 {
		return nil, errors.New("商品未上架")
	}

	// 校验SKU（可选）
	// skuID is optional pointer; if provided, validate it
	var skuPtr *uint
	if skuID != nil {
		var sku model.ProductSku
		if err := s.db.First(&sku, *skuID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("SKU不存在")
			}
			return nil, fmt.Errorf("获取SKU失败: %w", err)
		}
		if sku.ProductID != product.ID {
			return nil, errors.New("SKU与商品不匹配")
		}
		if sku.Status != 1 {
			return nil, errors.New("SKU未上架")
		}
		skuPtr = skuID
	}

	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// 查询是否已存在相同条目（处理 sku_id NULL 的情况）
	var item model.CartItem
	if skuPtr == nil {
		err = s.db.Where("cart_id = ? AND product_id = ? AND sku_id IS NULL", cart.ID, productID).First(&item).Error
	} else {
		err = s.db.Where("cart_id = ? AND product_id = ? AND sku_id = ?", cart.ID, productID, *skuID).First(&item).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = model.CartItem{CartID: cart.ID, ProductID: productID, SkuID: skuPtr, Quantity: quantity}
			if err := s.db.Create(&item).Error; err != nil {
				return nil, fmt.Errorf("添加到购物车失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("查询购物车条目失败: %w", err)
		}
	} else {
		// 已存在则累加数量
		newQty := item.Quantity + quantity
		if newQty <= 0 {
			newQty = 0
		}
		if err := s.db.Model(&item).Update("quantity", newQty).Error; err != nil {
			return nil, fmt.Errorf("更新购物车数量失败: %w", err)
		}
	}

	return &item, nil
}

// ListItems 获取购物车明细（附带商品/sku信息）
func (s *CartService) ListItems(userID uint) ([]model.CartItem, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}
	var items []model.CartItem
	if err := s.db.Where("cart_id = ?", cart.ID).Preload("Product").Preload("Sku").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("获取购物车列表失败: %w", err)
	}
	return items, nil
}

// GetStorePriceOverrides 返回购物车商品在指定门店的覆盖价映射。
// 目前仅针对门店特供（biz_type=3）的覆盖价。
func (s *CartService) GetStorePriceOverrides(storeID uint, items []model.CartItem) (map[uint]decimal.Decimal, error) {
	if storeID == 0 {
		return map[uint]decimal.Decimal{}, nil
	}
	if len(items) == 0 {
		return map[uint]decimal.Decimal{}, nil
	}

	ids := make([]uint, 0, len(items))
	seen := make(map[uint]struct{}, len(items))
	for _, it := range items {
		if it.ProductID == 0 {
			continue
		}
		if _, ok := seen[it.ProductID]; ok {
			continue
		}
		seen[it.ProductID] = struct{}{}
		ids = append(ids, it.ProductID)
	}
	if len(ids) == 0 {
		return map[uint]decimal.Decimal{}, nil
	}

	// biz_type=3 表示门店特供/其他
	const bizTypeExclusive = 3

	var rows []struct {
		ProductID     uint            `gorm:"column:product_id"`
		PriceOverride decimal.Decimal `gorm:"column:price_override"`
	}
	if err := s.db.Table("store_products").
		Select("product_id, price_override").
		Where("store_id = ? AND biz_type = ? AND product_id IN ?", storeID, bizTypeExclusive, ids).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("获取门店覆盖价失败: %w", err)
	}

	res := make(map[uint]decimal.Decimal, len(rows))
	for _, r := range rows {
		// 仅当覆盖价>0时才视为有效覆盖
		if r.ProductID == 0 {
			continue
		}
		if r.PriceOverride.GreaterThan(decimal.Zero) {
			res[r.ProductID] = r.PriceOverride
		}
	}
	return res, nil
}

// UpdateItem 更新购物车条目数量（<=0 则删除）
func (s *CartService) UpdateItem(userID uint, itemID uint, quantity int) error {
	// 验证条目归属
	var item model.CartItem
	if err := s.db.First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("购物车项不存在")
		}
		return fmt.Errorf("获取购物车项失败: %w", err)
	}

	var cart model.Cart
	if err := s.db.First(&cart, item.CartID).Error; err != nil {
		return fmt.Errorf("获取购物车失败: %w", err)
	}
	if cart.UserID != userID {
		return errors.New("无权操作该购物车项")
	}

	if quantity <= 0 {
		if err := s.db.Delete(&item).Error; err != nil {
			return fmt.Errorf("删除购物车项失败: %w", err)
		}
		return nil
	}

	if err := s.db.Model(&item).Update("quantity", quantity).Error; err != nil {
		return fmt.Errorf("更新购物车数量失败: %w", err)
	}
	return nil
}

// RemoveItem 删除购物车项
func (s *CartService) RemoveItem(userID uint, itemID uint) error {
	var item model.CartItem
	if err := s.db.First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("购物车项不存在")
		}
		return fmt.Errorf("获取购物车项失败: %w", err)
	}
	var cart model.Cart
	if err := s.db.First(&cart, item.CartID).Error; err != nil {
		return fmt.Errorf("获取购物车失败: %w", err)
	}
	if cart.UserID != userID {
		return errors.New("无权操作该购物车项")
	}
	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("删除购物车项失败: %w", err)
	}
	return nil
}

// Clear 清空购物车
func (s *CartService) Clear(userID uint) error {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return err
	}
	if err := s.db.Where("cart_id = ?", cart.ID).Delete(&model.CartItem{}).Error; err != nil {
		return fmt.Errorf("清空购物车失败: %w", err)
	}
	return nil
}
