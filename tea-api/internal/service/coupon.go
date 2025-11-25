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
