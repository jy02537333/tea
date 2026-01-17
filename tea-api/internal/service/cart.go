package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"

	"gorm.io/gorm/clause"
)

func normalizeStoreID(storeID *uint) *uint {
	if storeID == nil {
		return nil
	}
	if *storeID == 0 {
		return nil
	}
	sid := *storeID
	return &sid
}

func (s *CartService) getOrCreateCartTx(tx *gorm.DB, userID uint) (*model.Cart, error) {
	var cart model.Cart
	if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = model.Cart{UserID: userID}
			if err := tx.Create(&cart).Error; err != nil {
				return nil, fmt.Errorf("创建购物车失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("获取购物车失败: %w", err)
		}
	}
	return &cart, nil
}

// adjustReservedTx adjusts inventory by delta for a cart item.
// delta > 0: reserve (decrease available stock)
// delta < 0: release (increase available stock)
func (s *CartService) adjustReservedTx(tx *gorm.DB, product model.Product, sku *model.ProductSku, storeID *uint, delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		// Reserve store/platform stock
		if storeID != nil {
			resSp := tx.Model(&model.StoreProduct{}).
				Where("store_id = ? AND product_id = ? AND stock >= ?", *storeID, product.ID, delta).
				Update("stock", gorm.Expr("stock - ?", delta))
			if resSp.Error != nil {
				return fmt.Errorf("扣减门店库存失败: %w", resSp.Error)
			}
			if resSp.RowsAffected == 0 {
				return fmt.Errorf("门店库存不足: %s", product.Name)
			}
		} else {
			resP := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", product.ID, delta).
				Update("stock", gorm.Expr("stock - ?", delta))
			if resP.Error != nil {
				return fmt.Errorf("扣减商品库存失败: %w", resP.Error)
			}
			if resP.RowsAffected == 0 {
				return fmt.Errorf("商品库存不足: %s", product.Name)
			}
		}

		// Reserve SKU stock if applicable
		if sku != nil {
			resSku := tx.Model(&model.ProductSku{}).
				Where("id = ? AND stock >= ?", sku.ID, delta).
				Update("stock", gorm.Expr("stock - ?", delta))
			if resSku.Error != nil {
				return fmt.Errorf("扣减SKU库存失败: %w", resSku.Error)
			}
			if resSku.RowsAffected == 0 {
				return fmt.Errorf("SKU库存不足: %s", sku.SkuName)
			}
		}
		return nil
	}

	// Release
	release := -delta
	if storeID != nil {
		if err := tx.Model(&model.StoreProduct{}).
			Where("store_id = ? AND product_id = ?", *storeID, product.ID).
			Update("stock", gorm.Expr("stock + ?", release)).Error; err != nil {
			return fmt.Errorf("回补门店库存失败: %w", err)
		}
	} else {
		if err := tx.Model(&model.Product{}).
			Where("id = ?", product.ID).
			Update("stock", gorm.Expr("stock + ?", release)).Error; err != nil {
			return fmt.Errorf("回补商品库存失败: %w", err)
		}
	}
	if sku != nil {
		if err := tx.Model(&model.ProductSku{}).
			Where("id = ?", sku.ID).
			Update("stock", gorm.Expr("stock + ?", release)).Error; err != nil {
			return fmt.Errorf("回补SKU库存失败: %w", err)
		}
	}
	return nil
}

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
// storeID:
// - nil/0: 按平台商品库存 product.stock 校验
// - 非0: 按门店库存 store_products.stock 校验（门店维度）
func (s *CartService) AddItem(userID uint, productID uint, skuID *uint, quantity int, storeID *uint) (*model.CartItem, error) {
	if quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}
	storeID = normalizeStoreID(storeID)

	var out *model.CartItem
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 校验商品
		var product model.Product
		if err := tx.First(&product, productID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商品不存在")
			}
			return fmt.Errorf("获取商品失败: %w", err)
		}
		if product.Status != 1 {
			return errors.New("商品未上架")
		}

		// 门店维度必须有绑定记录
		if storeID != nil {
			var cnt int64
			if err := tx.Model(&model.StoreProduct{}).
				Where("store_id = ? AND product_id = ?", *storeID, productID).
				Count(&cnt).Error; err != nil {
				return fmt.Errorf("校验门店商品失败: %w", err)
			}
			if cnt == 0 {
				return fmt.Errorf("门店未上架该商品: %s", product.Name)
			}
		}

		// 校验SKU（可选）
		var skuPtr *uint
		var sku model.ProductSku
		var skuLoaded *model.ProductSku
		if skuID != nil {
			if err := tx.First(&sku, *skuID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("SKU不存在")
				}
				return fmt.Errorf("获取SKU失败: %w", err)
			}
			if sku.ProductID != product.ID {
				return errors.New("SKU与商品不匹配")
			}
			if sku.Status != 1 {
				return errors.New("SKU未上架")
			}
			skuPtr = skuID
			skuLoaded = &sku
		}

		cart, err := s.getOrCreateCartTx(tx, userID)
		if err != nil {
			return err
		}

		// 查询是否已存在相同条目（处理 sku_id NULL + store_id NULL 的情况）
		var item model.CartItem
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("cart_id = ? AND product_id = ?", cart.ID, productID)
		if skuPtr == nil {
			q = q.Where("sku_id IS NULL")
		} else {
			q = q.Where("sku_id = ?", *skuPtr)
		}
		if storeID == nil {
			q = q.Where("store_id IS NULL")
		} else {
			q = q.Where("store_id = ?", *storeID)
		}
		err = q.First(&item).Error

		found := true
		desiredQty := quantity
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				found = false
			} else {
				return fmt.Errorf("查询购物车条目失败: %w", err)
			}
		} else {
			desiredQty = item.Quantity + quantity
		}
		if desiredQty <= 0 {
			return errors.New("数量必须大于0")
		}

		// 计算需要增减的占用库存差值，并执行占用/回补
		delta := desiredQty
		reserved := 0
		if found {
			reserved = item.ReservedQuantity
			delta = desiredQty - reserved
		}
		if err := s.adjustReservedTx(tx, product, skuLoaded, storeID, delta); err != nil {
			return err
		}

		if !found {
			item = model.CartItem{
				CartID:           cart.ID,
				ProductID:        productID,
				SkuID:            skuPtr,
				StoreID:          storeID,
				Quantity:         desiredQty,
				ReservedQuantity: desiredQty,
			}
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("添加到购物车失败: %w", err)
			}
		} else {
			updates := map[string]any{
				"quantity":          desiredQty,
				"reserved_quantity": desiredQty,
				"updated_at":        time.Now(),
			}
			if err := tx.Model(&item).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新购物车数量失败: %w", err)
			}
		}

		out = &item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
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
// storeID:
// - nil/0: 按平台商品库存 product.stock 校验
// - 非0: 按门店库存 store_products.stock 校验（门店维度）
func (s *CartService) UpdateItem(userID uint, itemID uint, quantity int, storeID *uint) error {
	storeID = normalizeStoreID(storeID)
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
		return s.RemoveItem(userID, itemID)
	}

	// 如果请求侧传了 store_id，则与条目 store_id 要一致；否则按条目 store_id 执行。
	if storeID != nil {
		if item.StoreID == nil || *item.StoreID != *storeID {
			return errors.New("购物车条目门店不一致")
		}
	}
	storeCtx := item.StoreID

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 刷新商品/SKU，执行占用/回补差值
		var product model.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商品不存在")
			}
			return fmt.Errorf("获取商品失败: %w", err)
		}
		if product.Status != 1 {
			return errors.New("商品未上架")
		}
		var skuLoaded *model.ProductSku
		if item.SkuID != nil {
			var sku model.ProductSku
			if err := tx.First(&sku, *item.SkuID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("SKU不存在")
				}
				return fmt.Errorf("获取SKU失败: %w", err)
			}
			if sku.ProductID != product.ID {
				return errors.New("SKU与商品不匹配")
			}
			if sku.Status != 1 {
				return errors.New("SKU未上架")
			}
			skuLoaded = &sku
		}

		delta := quantity - item.ReservedQuantity
		if err := s.adjustReservedTx(tx, product, skuLoaded, storeCtx, delta); err != nil {
			return err
		}
		updates := map[string]any{
			"quantity":          quantity,
			"reserved_quantity": quantity,
			"updated_at":        time.Now(),
		}
		if err := tx.Model(&model.CartItem{}).Where("id = ?", item.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新购物车数量失败: %w", err)
		}
		return nil
	})
}

// RemoveItem 删除购物车项
func (s *CartService) RemoveItem(userID uint, itemID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item model.CartItem
		if err := tx.First(&item, itemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("购物车项不存在")
			}
			return fmt.Errorf("获取购物车项失败: %w", err)
		}
		var cart model.Cart
		if err := tx.First(&cart, item.CartID).Error; err != nil {
			return fmt.Errorf("获取购物车失败: %w", err)
		}
		if cart.UserID != userID {
			return errors.New("无权操作该购物车项")
		}

		// 回补占用库存
		if item.ReservedQuantity > 0 {
			var product model.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return fmt.Errorf("获取商品失败: %w", err)
			}
			var skuLoaded *model.ProductSku
			if item.SkuID != nil {
				var sku model.ProductSku
				if err := tx.First(&sku, *item.SkuID).Error; err != nil {
					return fmt.Errorf("获取SKU失败: %w", err)
				}
				skuLoaded = &sku
			}
			if err := s.adjustReservedTx(tx, product, skuLoaded, item.StoreID, -item.ReservedQuantity); err != nil {
				return err
			}
		}

		if err := tx.Delete(&item).Error; err != nil {
			return fmt.Errorf("删除购物车项失败: %w", err)
		}
		return nil
	})
}

// Clear 清空购物车
func (s *CartService) Clear(userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		cart, err := s.getOrCreateCartTx(tx, userID)
		if err != nil {
			return err
		}
		var items []model.CartItem
		if err := tx.Where("cart_id = ?", cart.ID).Find(&items).Error; err != nil {
			return fmt.Errorf("获取购物车条目失败: %w", err)
		}
		for _, it := range items {
			if it.ReservedQuantity <= 0 {
				continue
			}
			var product model.Product
			if err := tx.First(&product, it.ProductID).Error; err != nil {
				return fmt.Errorf("获取商品失败: %w", err)
			}
			var skuLoaded *model.ProductSku
			if it.SkuID != nil {
				var sku model.ProductSku
				if err := tx.First(&sku, *it.SkuID).Error; err != nil {
					return fmt.Errorf("获取SKU失败: %w", err)
				}
				skuLoaded = &sku
			}
			if err := s.adjustReservedTx(tx, product, skuLoaded, it.StoreID, -it.ReservedQuantity); err != nil {
				return err
			}
		}
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&model.CartItem{}).Error; err != nil {
			return fmt.Errorf("清空购物车失败: %w", err)
		}
		return nil
	})
}

// CleanupExpiredItems 清理超时购物车条目，并回补其占用库存。
// maxAge: 例如 30*time.Minute，表示 created_at <= now-maxAge 的条目会被清理。
// limit: 单次最多处理条数（<=0 表示不限制）。
func (s *CartService) CleanupExpiredItems(maxAge time.Duration, limit int) (int, error) {
	cutoff := time.Now().Add(-maxAge)
	processed := 0
	for {
		var batch []model.CartItem
		err := s.db.Transaction(func(tx *gorm.DB) error {
			q := tx.Where("created_at <= ?", cutoff).Order("id asc")
			if limit > 0 {
				q = q.Limit(limit)
			}
			if err := q.Find(&batch).Error; err != nil {
				return fmt.Errorf("查询超时购物车条目失败: %w", err)
			}
			if len(batch) == 0 {
				return nil
			}
			for _, it := range batch {
				// 行锁，避免和用户侧更新/删除冲突
				var locked model.CartItem
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, it.ID).Error; err != nil {
					return fmt.Errorf("锁定购物车条目失败: %w", err)
				}
				// 二次确认仍超时
				if locked.CreatedAt.After(cutoff) {
					continue
				}
				if locked.ReservedQuantity > 0 {
					var product model.Product
					if err := tx.First(&product, locked.ProductID).Error; err != nil {
						return fmt.Errorf("获取商品失败: %w", err)
					}
					var skuLoaded *model.ProductSku
					if locked.SkuID != nil {
						var sku model.ProductSku
						if err := tx.First(&sku, *locked.SkuID).Error; err != nil {
							return fmt.Errorf("获取SKU失败: %w", err)
						}
						skuLoaded = &sku
					}
					if err := s.adjustReservedTx(tx, product, skuLoaded, locked.StoreID, -locked.ReservedQuantity); err != nil {
						return err
					}
				}
				if err := tx.Delete(&locked).Error; err != nil {
					return fmt.Errorf("删除超时购物车条目失败: %w", err)
				}
				processed++
			}
			return nil
		})
		if err != nil {
			return processed, err
		}
		if len(batch) == 0 {
			break
		}
		// 如果没有 limit（或 limit 很大），避免无限循环：本轮已经处理完所有命中的条目
		if limit <= 0 {
			break
		}
	}
	return processed, nil
}
