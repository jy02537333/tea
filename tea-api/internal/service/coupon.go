package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type CouponService struct {
	db *gorm.DB
}

func NewCouponService() *CouponService { return &CouponService{db: database.GetDB()} }

func (s *CouponService) CreateCoupon(c *model.Coupon) error {
	if c.Name == "" {
		return errors.New("优惠券名称必填")
	}
	if c.Type < 1 || c.Type > 3 {
		return errors.New("非法的优惠券类型")
	}
	if c.Type == 1 && c.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("满减金额需大于0")
	}
	if c.Type == 2 && (c.Discount.LessThanOrEqual(decimal.Zero) || c.Discount.GreaterThan(decimal.NewFromInt(1))) {
		return errors.New("折扣需在(0,1]")
	}
	if c.TotalCount <= 0 {
		return errors.New("总发放数量需大于0")
	}
	if c.EndTime.Before(c.StartTime) {
		return errors.New("结束时间需大于开始时间")
	}
	return s.db.Create(c).Error
}

// CreateStoreCoupon 创建门店优惠券（在通用校验基础上强制绑定门店ID）
func (s *CouponService) CreateStoreCoupon(storeID uint, c *model.Coupon) error {
	if storeID == 0 {
		return errors.New("无效的门店ID")
	}
	c.StoreID = &storeID
	return s.CreateCoupon(c)
}

func (s *CouponService) ListCoupons(status int) ([]model.Coupon, error) {
	var list []model.Coupon
	q := s.db.Model(&model.Coupon{})
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListStoreCoupons 按门店列出优惠券
func (s *CouponService) ListStoreCoupons(storeID uint, status int) ([]model.Coupon, error) {
	var list []model.Coupon
	q := s.db.Model(&model.Coupon{}).Where("store_id = ?", storeID)
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateStoreCoupon 门店编辑自己的优惠券
func (s *CouponService) UpdateStoreCoupon(storeID, couponID uint, upd *model.Coupon) (*model.Coupon, error) {
	if storeID == 0 || couponID == 0 {
		return nil, errors.New("无效的参数")
	}
	var c model.Coupon
	if err := s.db.First(&c, couponID).Error; err != nil {
		return nil, err
	}
	if c.StoreID == nil || *c.StoreID != storeID {
		return nil, errors.New("无权操作该优惠券")
	}
	// 应用可编辑字段
	c.Name = upd.Name
	c.Type = upd.Type
	c.Amount = upd.Amount
	c.Discount = upd.Discount
	c.MinAmount = upd.MinAmount
	c.TotalCount = upd.TotalCount
	c.Status = upd.Status
	c.StartTime = upd.StartTime
	c.EndTime = upd.EndTime
	c.Description = upd.Description

	// 复用创建时的校验规则
	if c.Name == "" {
		return nil, errors.New("优惠券名称必填")
	}
	if c.Type < 1 || c.Type > 3 {
		return nil, errors.New("非法的优惠券类型")
	}
	if c.Type == 1 && c.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("满减金额需大于0")
	}
	if c.Type == 2 && (c.Discount.LessThanOrEqual(decimal.Zero) || c.Discount.GreaterThan(decimal.NewFromInt(1))) {
		return nil, errors.New("折扣需在(0,1]")
	}
	if c.TotalCount <= 0 {
		return nil, errors.New("总发放数量需大于0")
	}
	if c.EndTime.Before(c.StartTime) {
		return nil, errors.New("结束时间需大于开始时间")
	}
	if err := s.db.Save(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// GrantCouponToUser 发券给用户（简单版，不做并发扣减总量控制）
func (s *CouponService) GrantCouponToUser(couponID, userID uint) (*model.UserCoupon, error) {
	var c model.Coupon
	if err := s.db.First(&c, couponID).Error; err != nil {
		return nil, err
	}
	if c.Status != 1 {
		return nil, errors.New("优惠券未启用")
	}
	uc := &model.UserCoupon{UserID: userID, CouponID: c.ID, Status: 1}
	if err := s.db.Create(uc).Error; err != nil {
		return nil, err
	}
	return uc, nil
}

// ListUserAvailableCoupons 获取用户当前可用优惠券
func (s *CouponService) ListUserAvailableCoupons(userID uint) ([]model.UserCoupon, error) {
	now := time.Now()
	var list []model.UserCoupon
	if err := s.db.Preload("Coupon").
		Joins("JOIN coupons ON coupons.id = user_coupons.coupon_id").
		Where("user_coupons.user_id = ? AND user_coupons.status = 1 AND coupons.status = 1 AND ? BETWEEN coupons.start_time AND coupons.end_time", userID, now).
		Order("user_coupons.id desc").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
