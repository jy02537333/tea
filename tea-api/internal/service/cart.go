package service

import (
	"errors"
	"fmt"

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
func (s *CartService) AddItem(userID uint, productID, skuID uint, quantity int) (*model.CartItem, error) {
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
	if skuID != 0 {
		var sku model.ProductSku
		if err := s.db.First(&sku, skuID).Error; err != nil {
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
	}

	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// 查询是否已存在相同条目
	var item model.CartItem
	err = s.db.Where("cart_id = ? AND product_id = ? AND sku_id = ?", cart.ID, productID, skuID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = model.CartItem{CartID: cart.ID, ProductID: productID, SkuID: skuID, Quantity: quantity}
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
