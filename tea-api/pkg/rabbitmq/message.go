package rabbitmq

import (
	"encoding/json"
	"fmt"

	"tea-api/pkg/database"
)

// OrderMessage 订单消息结构
type OrderMessage struct {
	OrderID    uint   `json:"order_id"`
	UserID     uint   `json:"user_id"`
	Action     string `json:"action"` // created, paid, shipped, completed, cancelled
	TotalPrice int64  `json:"total_price"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
}

// PaymentMessage 支付消息结构
type PaymentMessage struct {
	PaymentID   uint   `json:"payment_id"`
	OrderID     uint   `json:"order_id"`
	UserID      uint   `json:"user_id"`
	Amount      int64  `json:"amount"`
	PaymentType string `json:"payment_type"` // wechat, alipay
	Status      string `json:"status"`       // success, failed, pending
	Timestamp   int64  `json:"timestamp"`
}

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	UserID    uint   `json:"user_id"`
	Type      string `json:"type"` // sms, wechat, push
	Title     string `json:"title"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// ExternalOrderMessage 外卖平台订单消息结构
type ExternalOrderMessage struct {
	ExternalOrderID string `json:"external_order_id"`
	Platform        string `json:"platform"` // meituan, eleme, baidu
	OrderID         uint   `json:"order_id"`
	Status          string `json:"status"`
	Action          string `json:"action"`
	Timestamp       int64  `json:"timestamp"`
}

// PublishOrderMessage 发布订单消息
func PublishOrderMessage(msg OrderMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化订单消息失败: %w", err)
	}

	routingKey := fmt.Sprintf("order.%s", msg.Action)
	return database.PublishMessage("tea_shop_exchange", routingKey, body)
}

// PublishPaymentMessage 发布支付消息
func PublishPaymentMessage(msg PaymentMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化支付消息失败: %w", err)
	}

	routingKey := fmt.Sprintf("payment.%s", msg.Status)
	return database.PublishMessage("tea_shop_exchange", routingKey, body)
}

// PublishNotificationMessage 发布通知消息
func PublishNotificationMessage(msg NotificationMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	routingKey := fmt.Sprintf("notification.%s", msg.Type)
	return database.PublishMessage("tea_shop_exchange", routingKey, body)
}

// PublishExternalOrderMessage 发布外卖平台订单消息
func PublishExternalOrderMessage(msg ExternalOrderMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化外卖订单消息失败: %w", err)
	}

	routingKey := fmt.Sprintf("external.order.%s.%s", msg.Platform, msg.Action)
	return database.PublishMessage("tea_shop_exchange", routingKey, body)
}
