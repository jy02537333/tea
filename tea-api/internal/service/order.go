package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService() *OrderService {
	return &OrderService{db: database.GetDB()}
}

// CreateMembershipOrder 为指定会员套餐创建一笔虚拟订单
// 该订单不依赖购物车，也不生成实体商品明细，仅用于会员/合伙人礼包购买场景。
// 约定：OrderType=4 表示会员订单，DeliveryType=1（自取/虚拟），StoreID=0。
func (s *OrderService) CreateMembershipOrder(userID, packageID uint, remark string) (*model.Order, error) {
	if userID == 0 || packageID == 0 {
		return nil, errors.New("非法的用户或套餐")
	}

	var pkg model.MembershipPackage
	if err := s.db.Where("id = ? AND type = ?", packageID, "membership").First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会员套餐不存在或已下架")
		}
		return nil, fmt.Errorf("查询会员套餐失败: %w", err)
	}

	// 暂不做重复购买限制与有效期校验，后续可在此扩展业务规则。
	order := &model.Order{
		OrderNo:             generateOrderNo("M"),
		UserID:              userID,
		StoreID:             0,
		MembershipPackageID: &packageID,
		Status:              1, // 待付款
		PayStatus:           1, // 未付款
		OrderType:           4, // 会员订单
		DeliveryType:        1, // 自取/虚拟
		AddressInfo:         "{}",
		Remark:              remark,
		TotalAmount:         pkg.Price,
		DiscountAmount:      decimal.NewFromInt(0),
		DeliveryFee:         decimal.NewFromInt(0),
		PayAmount:           pkg.Price,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, fmt.Errorf("创建会员订单失败: %w", err)
	}
	return order, nil
}

// CreateOrderFromCart 从购物车生成订单
func (s *OrderService) CreateOrderFromCart(userID uint, deliveryType int, addressInfo, remark string, userCouponID uint, storeID uint, orderType int) (*model.Order, error) {
	if deliveryType != 1 && deliveryType != 2 {
		return nil, errors.New("非法的配送类型")
	}
	if orderType == 0 {
		orderType = 1
	}
	if orderType != 1 && orderType != 2 && orderType != 3 {
		return nil, errors.New("非法的订单类型")
	}

	var cart model.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("购物车为空")
		}
		return nil, fmt.Errorf("获取购物车失败: %w", err)
	}

	var items []model.CartItem
	if err := s.db.Where("cart_id = ?", cart.ID).Preload("Product").Preload("Sku").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("获取购物车条目失败: %w", err)
	}
	if len(items) == 0 {
		return nil, errors.New("购物车为空")
	}

	order := &model.Order{
		OrderNo:        generateOrderNo("O"),
		UserID:         userID,
		StoreID:        storeID,
		Status:         1, // 待付款
		PayStatus:      1, // 未付款
		OrderType:      orderType,
		DeliveryType:   deliveryType,
		AddressInfo:    addressInfo,
		Remark:         remark,
		TotalAmount:    decimal.NewFromInt(0),
		DiscountAmount: decimal.NewFromInt(0),
		DeliveryFee:    decimal.NewFromInt(0),
		PayAmount:      decimal.NewFromInt(0),
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var couponIDForUse uint = 0
		// 校验门店（如传入）
		if storeID != 0 {
			var st model.Store
			if err := tx.First(&st, storeID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("门店不存在")
				}
				return err
			}
			if st.Status != 1 {
				return errors.New("门店不可用")
			}
		}

		// 逐项校验库存与价格，并扣减库存
		var orderItems []model.OrderItem
		total := decimal.NewFromInt(0)
		for _, it := range items {
			// 刷新商品/sku 以获取最新库存和价格
			var prod model.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&prod, it.ProductID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("商品不存在")
				}
				return fmt.Errorf("获取商品失败: %w", err)
			}
			if prod.Status != 1 {
				return fmt.Errorf("商品已下架: %s", prod.Name)
			}

			// 价格基于SKU优先
			price := prod.Price
			if it.SkuID != nil {
				var sku model.ProductSku
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sku, *it.SkuID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errors.New("SKU不存在")
					}
					return fmt.Errorf("获取SKU失败: %w", err)
				}
				if sku.Status != 1 {
					return fmt.Errorf("SKU未上架: %s", sku.SkuName)
				}
				if sku.ProductID != prod.ID {
					return errors.New("SKU与商品不匹配")
				}
				// 扣减 SKU 库存（乐观锁）
				res := tx.Model(&model.ProductSku{}).
					Where("id = ? AND stock >= ?", sku.ID, it.Quantity).
					Update("stock", gorm.Expr("stock - ?", it.Quantity))
				if res.Error != nil {
					return fmt.Errorf("扣减SKU库存失败: %w", res.Error)
				}
				if res.RowsAffected == 0 {
					return fmt.Errorf("SKU库存不足: %s", sku.SkuName)
				}
				price = sku.Price
			}

			// 如指定门店，则校验并扣减门店库存，并应用可能的门店价格覆盖
			if storeID != 0 {
				var sp model.StoreProduct
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("store_id = ? AND product_id = ?", storeID, it.ProductID).First(&sp).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return fmt.Errorf("门店未上架该商品: %s", prod.Name)
					}
					return err
				}
				resSp := tx.Model(&model.StoreProduct{}).
					Where("id = ? AND stock >= ?", sp.ID, it.Quantity).
					Update("stock", gorm.Expr("stock - ?", it.Quantity))
				if resSp.Error != nil {
					return fmt.Errorf("扣减门店库存失败: %w", resSp.Error)
				}
				if resSp.RowsAffected == 0 {
					return errors.New("门店库存不足")
				}
				if sp.PriceOverride.GreaterThan(decimal.Zero) {
					price = sp.PriceOverride
				}
			}

			// 扣减商品库存（乐观锁）
			res2 := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", prod.ID, it.Quantity).
				Update("stock", gorm.Expr("stock - ?", it.Quantity))
			if res2.Error != nil {
				return fmt.Errorf("扣减商品库存失败: %w", res2.Error)
			}
			if res2.RowsAffected == 0 {
				return fmt.Errorf("商品库存不足: %s", prod.Name)
			}

			qty := decimal.NewFromInt(int64(it.Quantity))
			amount := price.Mul(qty)
			total = total.Add(amount)

			var skuName string
			if it.SkuID != nil {
				skuName = it.Sku.SkuName
			}
			oi := model.OrderItem{
				ProductID:   prod.ID,
				SkuID:       it.SkuID,
				ProductName: prod.Name,
				SkuName:     skuName,
				Price:       price,
				Quantity:    it.Quantity,
				Amount:      amount,
				Image:       "",
			}
			orderItems = append(orderItems, oi)
		}

		order.TotalAmount = total

		// 应用优惠券（可选）
		discount := decimal.Zero
		if userCouponID != 0 {
			var uc model.UserCoupon
			// 兼容少量历史数据/写入差异：优先按 status=1 命中，失败则退化为按 (0,1) 命中
			if err := tx.Preload("Coupon").Where("id = ? AND user_id = ? AND status = 1", userCouponID, userID).First(&uc).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err2 := tx.Preload("Coupon").Where("id = ? AND user_id = ? AND status IN (0,1)", userCouponID, userID).First(&uc).Error; err2 != nil {
						if errors.Is(err2, gorm.ErrRecordNotFound) {
							return errors.New("无效的用户优惠券")
						}
						return err2
					}
				} else {
					return err
				}
			}
			if uc.Status == 2 || uc.Status == 3 {
				return errors.New("无效的用户优惠券")
			}
			nowT := time.Now()
			if uc.Coupon.Status != 1 || nowT.Before(uc.Coupon.StartTime) || nowT.After(uc.Coupon.EndTime) {
				return errors.New("优惠券不在有效期或已禁用")
			}
			if total.LessThan(uc.Coupon.MinAmount) {
				return errors.New("未满足优惠券使用门槛")
			}
			switch uc.Coupon.Type {
			case 1: // 满减券
				discount = uc.Coupon.Amount
			case 2: // 折扣券（0-1之间）
				one := decimal.NewFromInt(1)
				rate := uc.Coupon.Discount
				if rate.LessThanOrEqual(decimal.Zero) || rate.GreaterThan(one) {
					return errors.New("非法的折扣券配置")
				}
				discount = total.Mul(one.Sub(rate))
			case 3: // 免单券
				discount = total
			default:
				return errors.New("未知的优惠券类型")
			}
			if discount.GreaterThan(total) {
				discount = total
			}
			order.DiscountAmount = discount
			couponIDForUse = uc.CouponID
		}

		// 暂不收取运费
		order.PayAmount = total.Sub(discount)

		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 写入订单项
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return fmt.Errorf("创建订单明细失败: %w", err)
		}

		// 清空购物车
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&model.CartItem{}).Error; err != nil {
			return fmt.Errorf("清空购物车失败: %w", err)
		}

		// 标记优惠券为已使用并回写订单（如有）
		if userCouponID != 0 && order.DiscountAmount.GreaterThan(decimal.Zero) {
			usedAt := time.Now()
			if err := tx.Model(&model.UserCoupon{}).Where("id = ? AND user_id = ? AND status IN (0,1)", userCouponID, userID).
				Updates(map[string]any{"status": 2, "used_at": usedAt, "order_id": order.ID}).Error; err != nil {
				return fmt.Errorf("更新用户优惠券状态失败: %w", err)
			}
			// 累加优惠券使用次数
			if err := tx.Model(&model.Coupon{}).Where("id = ?", couponIDForUse).
				Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
				return fmt.Errorf("更新优惠券使用计数失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return order, nil
}

// ListOrders 列出用户订单
func (s *OrderService) ListOrders(userID uint, status int, page, limit int, storeID uint) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := s.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if storeID != 0 {
		q = q.Where("store_id = ?", storeID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []model.Order
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// GetOrder 获取订单详情（仅限本人）
func (s *OrderService) GetOrder(userID, orderID uint) (*model.Order, []model.OrderItem, error) {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("订单不存在")
		}
		return nil, nil, err
	}
	if order.UserID != userID {
		return nil, nil, errors.New("无权查看该订单")
	}
	var items []model.OrderItem
	if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &order, items, nil
}

// CancelOrder 取消订单（仅待付款可取消），并回补库存
func (s *OrderService) CancelOrder(userID, orderID uint, reason string) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	if order.UserID != userID {
		return errors.New("无权操作该订单")
	}
	if order.Status != 1 {
		return errors.New("当前状态不可取消")
	}

	var items []model.OrderItem
	if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		return err
	}
	// 回补库存（逐条更新，避免事务在 SQLite 上长时间锁表）
	for _, it := range items {
		if it.SkuID != nil {
			if err := s.db.Model(&model.ProductSku{}).Where("id = ?", *it.SkuID).
				Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
				return err
			}
		}
		if err := s.db.Model(&model.Product{}).Where("id = ?", it.ProductID).
			Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
			return err
		}
		// 如订单绑定了门店，回补门店库存
		if order.StoreID != 0 {
			if err := s.db.Model(&model.StoreProduct{}).
				Where("store_id = ? AND product_id = ?", order.StoreID, it.ProductID).
				Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
				return err
			}
		}
	}

	now := time.Now()
	order.Status = 5
	order.CancelledAt = &now
	order.CancelReason = reason
	if err := s.db.Save(&order).Error; err != nil {
		return err
	}
	return nil
}

// AdminCancelOrder 管理端取消订单（需权限），仅允许取消待付款订单；执行库存回补
func (s *OrderService) AdminCancelOrder(orderID uint, reason string) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	if order.Status != 1 {
		return errors.New("当前状态不可取消")
	}

	var items []model.OrderItem
	if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		return err
	}
	// 回补库存
	for _, it := range items {
		if it.SkuID != nil {
			if err := s.db.Model(&model.ProductSku{}).Where("id = ?", *it.SkuID).
				Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
				return err
			}
		}
		if err := s.db.Model(&model.Product{}).Where("id = ?", it.ProductID).
			Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
			return err
		}
		if order.StoreID != 0 {
			if err := s.db.Model(&model.StoreProduct{}).
				Where("store_id = ? AND product_id = ?", order.StoreID, it.ProductID).
				Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
				return err
			}
		}
	}

	now := time.Now()
	order.Status = 5
	order.CancelledAt = &now
	order.CancelReason = reason
	if err := s.db.Save(&order).Error; err != nil {
		return err
	}
	return nil
}

// AdminAdjustPayAmount 管理端调价（需权限）
// 规则：仅允许对待付款(Status=1)且未支付(PayStatus=1)的订单修改 PayAmount。
func (s *OrderService) AdminAdjustPayAmount(orderID uint, newPayAmount decimal.Decimal, reason string) error {
	_ = reason
 	if newPayAmount.LessThan(decimal.Zero) {
		return errors.New("金额不能为负")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("订单不存在")
			}
			return err
		}
		if order.Status != 1 || order.PayStatus != 1 {
			return errors.New("仅支持未支付的待付款订单调价")
		}

		order.PayAmount = newPayAmount
		return tx.Save(&order).Error
	})
}

func generateOrderNo(prefix string) string {
	// 简单生成：前缀 + 时间戳 + 随机（取 UID 的前缀）
	ts := time.Now().Format("20060102150405")
	uid := utils.GenerateUID()
	if len(uid) > 6 {
		uid = uid[:6]
	}
	return fmt.Sprintf("%s%s%s", prefix, ts, uid)
}

// MarkPaid 模拟支付成功：仅待付款可支付
func (s *OrderService) MarkPaid(userID, orderID uint) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	// 支付只能由订单所属用户执行
	if order.UserID != userID {
		return errors.New("无权操作该订单")
	}
	if order.Status != 1 || order.PayStatus != 1 {
		return errors.New("当前状态不可支付")
	}
	now := time.Now()
	order.Status = 2    // 已付款
	order.PayStatus = 2 // 已付款
	order.PaidAt = &now
	return s.db.Save(&order).Error
}

// StartDelivery 发货：仅已付款可发货
func (s *OrderService) StartDelivery(userID, orderID uint) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	// 发货由具有相应权限的用户执行，不强制订单归属校验（权限由中间件控制）
	if order.Status != 2 {
		return errors.New("当前状态不可发货")
	}
	order.Status = 3 // 配送中
	return s.db.Save(&order).Error
}

// Complete 完成订单：仅配送中可完成
func (s *OrderService) Complete(userID, orderID uint) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	// 完成由具有相应权限的用户执行，不强制订单归属校验（权限由中间件控制）
	if order.Status != 3 {
		return errors.New("当前状态不可完成")
	}
	now := time.Now()
	order.Status = 4
	order.CompletedAt = &now
	return s.db.Save(&order).Error
}

// Receive 用户确认收货/完成订单：
// - 配送(DeliveryType=2)：仅当状态为配送中(3)可确认
// - 自取(DeliveryType=1)：仅当状态为已付款(2)可确认
// 仅允许订单所属用户操作
func (s *OrderService) Receive(userID, orderID uint) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	if order.UserID != userID {
		return errors.New("无权操作该订单")
	}
	switch order.DeliveryType {
	case 2: // 配送
		if order.Status != 3 {
			return errors.New("当前状态不可确认收货")
		}
	case 1: // 自取
		if order.Status != 2 {
			return errors.New("当前状态不可确认收货")
		}
	default:
		return errors.New("非法的配送类型")
	}
	now := time.Now()
	order.Status = 4
	order.CompletedAt = &now
	return s.db.Save(&order).Error
}

// AdminRefundOrder 管理端手动退款（需权限）
// 规则：
// - 仅在 PayStatus=2(已付款) 时可退款
// - 允许状态为 已付款(2) 或 配送中(3)
// - 若状态为 已付款(2)（未发货），执行库存回补；若已配送中(3)，不回补库存
// - 将订单状态置为 已取消(5)，支付状态置为 已退款(4)
func (s *OrderService) AdminRefundOrder(orderID uint, reason string) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	if order.PayStatus != 2 {
		return errors.New("当前支付状态不可退款")
	}
	if order.Status != 2 && order.Status != 3 {
		return errors.New("当前状态不可退款")
	}

	if order.Status == 2 { // 未发货，回补库存
		var items []model.OrderItem
		if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			return err
		}
		for _, it := range items {
			if it.SkuID != nil {
				if err := s.db.Model(&model.ProductSku{}).Where("id = ?", *it.SkuID).
					Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
					return err
				}
			}
			if err := s.db.Model(&model.Product{}).Where("id = ?", it.ProductID).
				Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
				return err
			}
			if order.StoreID != 0 {
				if err := s.db.Model(&model.StoreProduct{}).
					Where("store_id = ? AND product_id = ?", order.StoreID, it.ProductID).
					Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
					return err
				}
			}
		}
	}

	// 标记为已退款+已取消
	now := time.Now()
	order.Status = 5
	order.PayStatus = 4
	order.CancelledAt = &now
	if reason != "" {
		order.CancelReason = reason
	}
	if err := s.db.Save(&order).Error; err != nil {
		return err
	}

	// 回滚已使用的优惠券（如有）
	var uc model.UserCoupon
	if err := s.db.Where("order_id = ? AND status = 2", order.ID).First(&uc).Error; err == nil {
		if err2 := s.db.Model(&model.UserCoupon{}).Where("id = ?", uc.ID).
			Updates(map[string]any{"status": 1, "used_at": nil, "order_id": nil}).Error; err2 != nil {
			return err2
		}
		// 安全递减已使用计数
		if err2 := s.db.Model(&model.Coupon{}).Where("id = ? AND used_count > 0", uc.CouponID).
			Update("used_count", gorm.Expr("used_count - 1")).Error; err2 != nil {
			return err2
		}
	}
	return nil
}

// AdminRefundStart 标记退款中
func (s *OrderService) AdminRefundStart(orderID uint, reason string) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	if order.PayStatus != 2 {
		return errors.New("当前支付状态不可标记退款")
	}
	if order.Status != 2 && order.Status != 3 {
		return errors.New("当前状态不可标记退款")
	}
	order.PayStatus = 3 // 退款中
	if reason != "" {
		order.CancelReason = reason
	}
	return s.db.Save(&order).Error
}

// AdminRefundConfirm 确认退款完成
func (s *OrderService) AdminRefundConfirm(orderID uint, reason string) error {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return err
	}
	if order.PayStatus != 3 {
		return errors.New("未处于退款中状态")
	}
	if order.Status != 2 && order.Status != 3 {
		return errors.New("当前状态不可确认退款")
	}

	// 若未发货则回补库存
	if order.Status == 2 {
		var items []model.OrderItem
		if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			return err
		}
		for _, it := range items {
			if it.SkuID != nil {
				if err := s.db.Model(&model.ProductSku{}).Where("id = ?", *it.SkuID).
					Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
					return err
				}
			}
			if err := s.db.Model(&model.Product{}).Where("id = ?", it.ProductID).
				Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
				return err
			}
			if order.StoreID != 0 {
				if err := s.db.Model(&model.StoreProduct{}).
					Where("store_id = ? AND product_id = ?", order.StoreID, it.ProductID).
					Update("stock", gorm.Expr("stock + ?", it.Quantity)).Error; err != nil {
					return err
				}
			}
		}
	}

	// 完成退款
	now := time.Now()
	order.Status = 5
	order.PayStatus = 4
	order.CancelledAt = &now
	if reason != "" {
		order.CancelReason = reason
	}
	if err := s.db.Save(&order).Error; err != nil {
		return err
	}
	// 回滚优惠券
	var uc model.UserCoupon
	if err := s.db.Where("order_id = ? AND status = 2", order.ID).First(&uc).Error; err == nil {
		if err2 := s.db.Model(&model.UserCoupon{}).Where("id = ?", uc.ID).
			Updates(map[string]any{"status": 1, "used_at": nil, "order_id": nil}).Error; err2 != nil {
			return err2
		}
		if err2 := s.db.Model(&model.Coupon{}).Where("id = ? AND used_count > 0", uc.CouponID).
			Update("used_count", gorm.Expr("used_count - 1")).Error; err2 != nil {
			return err2
		}
	}
	return nil
}

// StoreOrderStats 门店订单统计结果
type StoreOrderStats struct {
	StoreID         uint              `json:"store_id"`
	TotalOrders     int64             `json:"total_orders"`
	CompletedAmount decimal.Decimal   `json:"completed_amount"`
	StatusCounts    []StatusCountItem `json:"status_counts"`
}

type StatusCountItem struct {
	Status int   `json:"status"`
	Count  int64 `json:"count"`
}

// GetStoreOrderStats 获取指定门店的订单统计（成交额按已完成订单汇总）
func (s *OrderService) GetStoreOrderStats(storeID uint) (*StoreOrderStats, error) {
	if storeID == 0 {
		return nil, errors.New("store_id 不能为空")
	}
	// 总订单数
	var total int64
	if err := s.db.Model(&model.Order{}).Where("store_id = ?", storeID).Count(&total).Error; err != nil {
		return nil, err
	}
	// 成交额（已完成）
	var completedAmount decimal.Decimal
	row := s.db.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount), 0)").
		Where("store_id = ? AND status = 4", storeID).
		Row()
	if err := row.Scan(&completedAmount); err != nil {
		return nil, err
	}
	// 各状态计数
	type sc struct {
		Status int
		Count  int64
	}
	var rows []sc
	if err := s.db.Model(&model.Order{}).
		Select("status, COUNT(*) as count").
		Where("store_id = ?", storeID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	var statusItems []StatusCountItem
	for _, r := range rows {
		statusItems = append(statusItems, StatusCountItem{Status: r.Status, Count: r.Count})
	}
	return &StoreOrderStats{StoreID: storeID, TotalOrders: total, CompletedAmount: completedAmount, StatusCounts: statusItems}, nil
}

// AdminListOrders 列出所有订单（管理端），支持按 store_id、status、时间区间过滤
func (s *OrderService) AdminListOrders(status int, page, limit int, storeID uint, startTime, endTime *time.Time) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	q := s.db.Model(&model.Order{})
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if storeID != 0 {
		q = q.Where("store_id = ?", storeID)
	}
	if startTime != nil {
		q = q.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		q = q.Where("created_at <= ?", *endTime)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []model.Order
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// GetOrderAdmin 获取任意订单详情（管理端），不校验 user_id
func (s *OrderService) GetOrderAdmin(orderID uint) (*model.Order, []model.OrderItem, error) {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("订单不存在")
		}
		return nil, nil, err
	}
	var items []model.OrderItem
	if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &order, items, nil
}

// AdminListStoreOrders 按门店维度列出订单列表，支持状态、时间区间和订单ID筛选
// - storeID: 必填，门店 ID
// - status: 可选，按订单状态过滤（>0 时生效）
// - page/limit: 分页参数；limit 超出合理范围时回退到 100
// - startTime/endTime: 可选，按创建时间区间过滤
// - orderID: 可选，按订单主键精确过滤
func (s *OrderService) AdminListStoreOrders(storeID uint, status int, page, limit int, startTime, endTime *time.Time, orderID uint) ([]model.Order, int64, error) {
	if storeID == 0 {
		return nil, 0, errors.New("store_id 不能为空")
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	q := s.db.Model(&model.Order{}).Where("store_id = ?", storeID)
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if startTime != nil {
		q = q.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		q = q.Where("created_at <= ?", *endTime)
	}
	if orderID > 0 {
		q = q.Where("id = ?", orderID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []model.Order
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
