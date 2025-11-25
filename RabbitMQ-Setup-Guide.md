# RabbitMQ 配置和使用指南

## 📋 已配置的RabbitMQ设置

### 连接信息
- **服务器**: 10.8.0.14
- **端口**: 5672
- **用户名**: guest
- **密码**: guest
- **虚拟主机**: /

### 消息队列架构

#### 交换机 (Exchange)
- **名称**: `tea_shop_exchange`
- **类型**: topic
- **持久化**: 是

#### 队列 (Queues)
1. **订单队列**: `tea_shop_order_queue`
   - 路由键: `order.*`
   - 处理: 订单创建、支付、发货、完成、取消等

2. **支付队列**: `tea_shop_payment_queue`
   - 路由键: `payment.*`
   - 处理: 支付成功、失败、退款等

3. **通知队列**: `tea_shop_notification_queue`
   - 路由键: `notification.*`
   - 处理: 短信、微信、推送通知等

4. **外卖平台队列**: `tea_shop_external_order_queue`
   - 路由键: `external.order.*`
   - 处理: 美团、饿了么、百度外卖订单同步

## 🚀 消息发布示例

### 订单消息
```go
import "tea-api/pkg/rabbitmq"

// 发布订单创建消息
orderMsg := rabbitmq.OrderMessage{
    OrderID:    123,
    UserID:     456,
    Action:     "created",
    TotalPrice: 9800, // 98.00元，以分为单位
    Status:     "pending",
    Timestamp:  time.Now().Unix(),
}
err := rabbitmq.PublishOrderMessage(orderMsg)
```

### 支付消息
```go
// 发布支付成功消息
paymentMsg := rabbitmq.PaymentMessage{
    PaymentID:   789,
    OrderID:     123,
    UserID:      456,
    Amount:      9800,
    PaymentType: "wechat",
    Status:      "success",
    Timestamp:   time.Now().Unix(),
}
err := rabbitmq.PublishPaymentMessage(paymentMsg)
```

### 通知消息
```go
// 发布微信通知消息
notificationMsg := rabbitmq.NotificationMessage{
    UserID:    456,
    Type:      "wechat",
    Title:     "订单支付成功",
    Content:   "您的订单已支付成功，我们将尽快为您配送",
    Timestamp: time.Now().Unix(),
}
err := rabbitmq.PublishNotificationMessage(notificationMsg)
```

### 外卖平台订单消息
```go
// 发布美团订单同步消息
externalMsg := rabbitmq.ExternalOrderMessage{
    ExternalOrderID: "MT202511060001",
    Platform:        "meituan",
    OrderID:         123,
    Status:          "accepted",
    Action:          "sync",
    Timestamp:       time.Now().Unix(),
}
err := rabbitmq.PublishExternalOrderMessage(externalMsg)
```

## 🔧 RabbitMQ 服务确认

### 检查RabbitMQ服务状态
确保RabbitMQ服务在10.8.0.14:5672上运行：

```bash
# 检查端口是否开放
telnet 10.8.0.14 5672

# 或使用PowerShell测试连接
Test-NetConnection -ComputerName 10.8.0.14 -Port 5672
```

### RabbitMQ管理界面
如果启用了管理插件，可以访问：
- **URL**: http://10.8.0.14:15672
- **用户名**: guest
- **密码**: guest

## ⚡ 容错处理

当前配置中，如果RabbitMQ连接失败：
- ✅ 服务器会继续启动，不会崩溃
- ⚠️ 消息队列功能将不可用
- 📝 会输出详细的错误信息

### 启动日志示例

**RabbitMQ连接成功时：**
```
正在连接RabbitMQ: 10.8.0.14:5672
RabbitMQ连接字符串: amqp://guest:***@10.8.0.14:5672/
RabbitMQ连接成功!
RabbitMQ通道创建成功!
RabbitMQ交换机和队列初始化成功!
```

**RabbitMQ连接失败时：**
```
正在连接RabbitMQ: 10.8.0.14:5672
RabbitMQ连接失败，但继续启动服务器: dial tcp 10.8.0.14:5672: connect: connection refused
注意：RabbitMQ功能将不可用，请确保RabbitMQ服务已启动
```

## 🎯 在业务代码中使用

### 订单服务示例
```go
// 在订单创建后发布消息
func CreateOrder(order *model.Order) error {
    // 保存订单到数据库
    if err := db.Create(order).Error; err != nil {
        return err
    }
    
    // 发布订单创建消息到RabbitMQ
    orderMsg := rabbitmq.OrderMessage{
        OrderID:    order.ID,
        UserID:     order.UserID,
        Action:     "created",
        TotalPrice: order.TotalPrice,
        Status:     order.Status,
        Timestamp:  time.Now().Unix(),
    }
    
    // 异步发送，不阻塞主流程
    go func() {
        if err := rabbitmq.PublishOrderMessage(orderMsg); err != nil {
            log.Printf("发布订单消息失败: %v", err)
        }
    }()
    
    return nil
}
```

## 📊 队列监控

推荐监控以下指标：
- 队列消息数量
- 消费速率
- 错误率
- 连接状态

这样可以确保消息队列系统的健康运行。