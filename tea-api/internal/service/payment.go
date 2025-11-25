package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type PaymentService struct{ db *gorm.DB }

func NewPaymentService() *PaymentService { return &PaymentService{db: database.GetDB()} }

// CreateIntent 创建支付意图（模拟）
func (s *PaymentService) CreateIntent(userID, orderID uint, method int) (*model.Payment, string, error) {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("订单不存在")
		}
		return nil, "", err
	}
	if order.UserID != userID {
		return nil, "", errors.New("无权为该订单创建支付")
	}
	if order.Status != 1 || order.PayStatus != 1 {
		return nil, "", errors.New("订单当前不可创建支付")
	}

	pay := &model.Payment{
		OrderID:       order.ID,
		PaymentNo:     generatePaymentNo("P"),
		PaymentMethod: method,
		Amount:        order.PayAmount,
		Status:        1, // 待支付
	}
	if err := s.db.Create(pay).Error; err != nil {
		return nil, "", err
	}
	// 模拟支付URL
	payURL := fmt.Sprintf("mockpay://%s", pay.PaymentNo)
	return pay, payURL, nil
}

// MockCallback 模拟第三方回调，根据 payment_no 标记支付成功并更新订单
func (s *PaymentService) MockCallback(paymentNo string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var pay model.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("payment_no = ?", paymentNo).First(&pay).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("支付记录不存在")
			}
			return err
		}
		if pay.Status == 2 { // 已成功
			return nil
		}
		var order model.Order
		if err := tx.First(&order, pay.OrderID).Error; err != nil {
			return err
		}

		now := time.Now()
		pay.Status = 2
		pay.PaidAt = &now
		if err := tx.Save(&pay).Error; err != nil {
			return err
		}

		// 标记订单已支付
		// 直接设置（模拟回调走系统身份）
		order.Status = 2
		order.PayStatus = 2
		order.PaidAt = &now
		return tx.Save(&order).Error
	})
}

func generatePaymentNo(prefix string) string {
	ts := time.Now().Format("20060102150405")
	// 使用纳秒避免并发冲突
	return fmt.Sprintf("%s%s%09d", prefix, ts, time.Now().UnixNano()%1e9)
}

// helpers (kept for completeness)
var _ = decimal.NewFromInt // silence unused import if optimized by compiler
