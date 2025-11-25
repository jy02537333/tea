package database

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"tea-api/internal/config"
)

var RMQ *amqp.Connection
var RMQChannel *amqp.Channel

// InitRabbitMQ 初始化RabbitMQ连接
func InitRabbitMQ() {
	cfg := config.Config.RabbitMQ

	fmt.Printf("正在连接RabbitMQ: %s:%d\n", cfg.Host, cfg.Port)

	// 建立连接
	dsn := cfg.DSN()
	fmt.Printf("RabbitMQ连接字符串: %s\n", dsn)

	conn, err := amqp.Dial(dsn)
	if err != nil {
		fmt.Printf("RabbitMQ连接失败，但继续启动服务器: %v\n", err)
		fmt.Println("注意：RabbitMQ功能将不可用，请确保RabbitMQ服务已启动")
		RMQ = nil
		RMQChannel = nil
		return
		// 如果需要强制要求RabbitMQ，可以取消下面这行的注释
		// panic(fmt.Errorf("failed to connect rabbitmq: %w", err))
	}

	RMQ = conn
	fmt.Println("RabbitMQ连接成功!")

	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("RabbitMQ通道创建失败: %v\n", err)
		RMQChannel = nil
		return
	}

	RMQChannel = ch
	fmt.Println("RabbitMQ通道创建成功!")

	// 初始化交换机和队列
	if err := initExchangeAndQueues(); err != nil {
		fmt.Printf("初始化RabbitMQ交换机和队列失败: %v\n", err)
	} else {
		fmt.Println("RabbitMQ交换机和队列初始化成功!")
	}
}

// initExchangeAndQueues 初始化交换机和队列
func initExchangeAndQueues() error {
	if RMQChannel == nil {
		return fmt.Errorf("RabbitMQ通道未初始化")
	}

	cfg := config.Config.RabbitMQ

	// 声明交换机
	err := RMQChannel.ExchangeDeclare(
		cfg.Exchange.TeaShop, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return fmt.Errorf("声明交换机失败: %w", err)
	}

	// 声明队列
	queues := []struct {
		name       string
		routingKey string
	}{
		{cfg.Queues.Order, "order.*"},
		{cfg.Queues.Payment, "payment.*"},
		{cfg.Queues.Notification, "notification.*"},
		{cfg.Queues.ExternalOrder, "external.order.*"},
	}

	for _, queue := range queues {
		// 声明队列
		_, err := RMQChannel.QueueDeclare(
			queue.name, // name
			true,       // durable
			false,      // delete when unused
			false,      // exclusive
			false,      // no-wait
			nil,        // arguments
		)
		if err != nil {
			return fmt.Errorf("声明队列 %s 失败: %w", queue.name, err)
		}

		// 绑定队列到交换机
		err = RMQChannel.QueueBind(
			queue.name,           // queue name
			queue.routingKey,     // routing key
			cfg.Exchange.TeaShop, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("绑定队列 %s 失败: %w", queue.name, err)
		}
	}

	return nil
}

// GetRabbitMQ 获取RabbitMQ连接
func GetRabbitMQ() *amqp.Connection {
	return RMQ
}

// GetRabbitMQChannel 获取RabbitMQ通道
func GetRabbitMQChannel() *amqp.Channel {
	return RMQChannel
}

// PublishMessage 发布消息到RabbitMQ
func PublishMessage(exchange, routingKey string, body []byte) error {
	if RMQChannel == nil {
		return fmt.Errorf("RabbitMQ通道未初始化")
	}

	return RMQChannel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)
}

// CloseRabbitMQ 关闭RabbitMQ连接
func CloseRabbitMQ() {
	if RMQChannel != nil {
		RMQChannel.Close()
	}
	if RMQ != nil {
		RMQ.Close()
	}
}
