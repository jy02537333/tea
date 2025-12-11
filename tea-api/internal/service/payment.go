package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type PaymentService struct{ db *gorm.DB }

func NewPaymentService() *PaymentService { return &PaymentService{db: database.GetDB()} }

// UnifiedOrderResult 描述统一下单结果，供前端/小程序拉起支付
type UnifiedOrderResult struct {
	PaymentNo string          `json:"payment_no"`
	OrderID   uint            `json:"order_id"`
	Method    int             `json:"method"`
	Amount    decimal.Decimal `json:"amount"`
	PrepayID  string          `json:"prepay_id"`
	NonceStr  string          `json:"nonce_str"`
	Timestamp int64           `json:"timestamp"`
	Package   string          `json:"package"`
	Sign      string          `json:"sign"`
	PayURL    string          `json:"pay_url"`
}

// PaymentCallbackPayload 为第三方支付回调封装的参数
type PaymentCallbackPayload struct {
	PaymentNo     string
	TransactionID string
	TradeState    string
	Sign          string
	PaidAt        *time.Time
	RawBody       string
	TestMode      bool
	SkipVerify    bool
}

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

	var existing model.Payment
	if err := s.db.Where("order_id = ? AND payment_method = ? AND status = 1", order.ID, method).
		Order("id desc").First(&existing).Error; err == nil {
		return &existing, fmt.Sprintf("mockpay://%s", existing.PaymentNo), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
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

// UnifiedOrder 模拟微信统一下单，返回前端拉起支付所需参数
func (s *PaymentService) UnifiedOrder(userID, orderID uint, method int) (*UnifiedOrderResult, error) {
	pay, payURL, err := s.CreateIntent(userID, orderID, method)
	if err != nil {
		return nil, err
	}
	nonce := utils.GenerateRandomString(18)
	timestamp := time.Now().Unix()
	prepayID := fmt.Sprintf("prepay_%s", pay.PaymentNo)
	packageStr := "Sign=WXPay"
	secret := config.Config.WeChat.APIKey
	signPayload := fmt.Sprintf("%s|%s|%d|%s", prepayID, nonce, timestamp, secret)
	sign := strings.ToUpper(utils.MD5Hash(signPayload))
	return &UnifiedOrderResult{
		PaymentNo: pay.PaymentNo,
		OrderID:   pay.OrderID,
		Method:    pay.PaymentMethod,
		Amount:    pay.Amount,
		PrepayID:  prepayID,
		NonceStr:  nonce,
		Timestamp: timestamp,
		Package:   packageStr,
		Sign:      sign,
		PayURL:    payURL,
	}, nil
}

// MockCallback 模拟第三方回调，根据 payment_no 标记支付成功并更新订单
func (s *PaymentService) MockCallback(paymentNo string) error {
	payload := PaymentCallbackPayload{
		PaymentNo:  paymentNo,
		TradeState: "SUCCESS",
		SkipVerify: true,
		RawBody:    fmt.Sprintf("{\"mock\":true,\"payment_no\":\"%s\"}", paymentNo),
	}
	return s.HandleCallback(payload)
}

// HandleCallback 处理第三方支付回调
func (s *PaymentService) HandleCallback(payload PaymentCallbackPayload) error {
	if payload.PaymentNo == "" {
		return errors.New("payment_no 不能为空")
	}
	if payload.TradeState == "" {
		return errors.New("trade_state 不能为空")
	}
	if !payload.SkipVerify {
		secret := config.Config.WeChat.APIKey
		expected := strings.ToUpper(utils.MD5Hash(fmt.Sprintf("%s|%s|%s", payload.PaymentNo, payload.TradeState, secret)))
		if payload.Sign == "" || payload.Sign != expected {
			return errors.New("签名校验失败")
		}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var pay model.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("payment_no = ?", payload.PaymentNo).First(&pay).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("支付记录不存在")
			}
			return err
		}

		now := time.Now()
		pay.ThirdPayNo = payload.TransactionID
		pay.ThirdResponse = payload.RawBody
		pay.NotifyAt = &now

		if payload.TradeState != "SUCCESS" {
			pay.Status = 3
			if err := tx.Save(&pay).Error; err != nil {
				return err
			}
			return nil
		}

		if pay.Status == 2 {
			return nil
		}

		paidAt := payload.PaidAt
		if paidAt == nil {
			paidAt = &now
		}
		pay.Status = 2
		pay.PaidAt = paidAt
		if err := tx.Save(&pay).Error; err != nil {
			return err
		}

		var order model.Order
		if err := tx.First(&order, pay.OrderID).Error; err != nil {
			return err
		}
		order.Status = 2
		order.PayStatus = 2
		order.PaidAt = paidAt
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// 若该订单关联活动报名记录，则将报名状态从「已报名」更新为「已支付报名」
		if err := tx.Model(&model.ActivityRegistration{}).
			Where("order_id = ? AND status = ?", order.ID, 1).
			Update("status", 2).Error; err != nil {
			return err
		}
		return nil
	})
}

func generatePaymentNo(prefix string) string {
	ts := time.Now().Format("20060102150405")
	// 使用纳秒避免并发冲突
	return fmt.Sprintf("%s%s%09d", prefix, ts, time.Now().UnixNano()%1e9)
}

// helpers (kept for completeness)
var _ = decimal.NewFromInt // silence unused import if optimized by compiler

// MarshalUnifiedOrderResult 方便日志输出
func MarshalUnifiedOrderResult(res *UnifiedOrderResult) string {
	if res == nil {
		return "{}"
	}
	b, _ := json.Marshal(res)
	return string(b)
}
