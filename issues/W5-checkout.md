# W5 结算 / 下单（Checkout / Orders）

**目标**：下单流程（地址/自取、优惠券、下单、调用支付 intent）

**路由**：`/pages/checkout`

**关键接口**：`POST /api/v1/orders/from-cart`、`POST /api/v1/payment/intents`

**验收标准**：下单成功后跳转 order-detail 并展示支付入口或已支付状态

**估时**：2d
