# 订单模块 API 文档

## 基础约定

- Base URL：`/api/v1`
- 返回格式统一为：

```json
{
  "code": 0,
  "data": "...",
  "message": "ok"
}
```

- 需要登录的接口使用 JWT Bearer：
  - Header: `Authorization: Bearer <token>`

---

## 一、用户侧订单 API

### 1. POST `/api/v1/orders/from-cart` 从购物车创建订单

- 鉴权：需要登录（`AuthMiddleware`）
- 请求体：

```json
{
  "delivery_type": 1,
  "address_info": "JSON 字符串，可为空",
  "remark": "备注",
  "user_coupon_id": 0,
  "store_id": 0,
  "order_type": 1
}
```

- 字段说明：
  - `delivery_type`：必填，`1` 自取，`2` 配送。
  - `address_info`：字符串，内部作为 JSON 存储，留给前端自定义结构（收货地址等）。
  - `remark`：字符串，用户备注。
  - `user_coupon_id`：可选，用户优惠券 ID。
  - `store_id`：可选，门店 ID（支持门店维度订单）。
  - `order_type`：`1` 商城，`2` 堂食，`3` 外卖。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "id": 123,
    "order_no": "202512080001",
    "pay_amount": 20.0,
    "discount_amount": 0.0
  },
  "message": "ok"
}
```

---

### 2. GET `/api/v1/orders` 用户订单列表

- 鉴权：需要登录
- 查询参数：
  - `page`：页码，默认 `1`
  - `limit`：每页数量，默认 `20`
  - `status`：订单状态，可选，默认 `0` 表示不过滤。
    - 1 待付款，2 已付款，3 配送中，4 已完成，5 已取消
  - `store_id`：门店 ID，可选，支持过滤门店订单。

- 响应示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": 123,
      "order_no": "202512080001",
      "user_id": 1,
      "store_id": 0,
      "total_amount": "20.00",
      "pay_amount": "20.00",
      "discount_amount": "0.00",
      "delivery_fee": "0.00",
      "status": 1,
      "pay_status": 1,
      "order_type": 1,
      "delivery_type": 1,
      "delivery_time": null,
      "address_info": "{}",
      "remark": "",
      "paid_at": null,
      "delivered_at": null,
      "completed_at": null,
      "cancelled_at": null,
      "cancel_reason": "",
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 1
}
```

> 列表中返回的是 `Order` 模型的 JSON：金额字段是字符串（decimal 转字符串）。

---

### 3. GET `/api/v1/orders/:id` 用户订单详情

- 鉴权：需要登录（只允许本人查看）。
- 路径参数：
  - `id`：订单 ID（数字）。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "order": {
      "...": "同列表 Order 字段"
    },
    "items": [
      {
        "id": 1,
        "order_id": 123,
        "product_id": 10,
        "sku_id": null,
        "product_name": "下单商品",
        "sku_name": "",
        "price": "10.00",
        "quantity": 2,
        "amount": "20.00",
        "image": ""
      }
    ]
  },
  "message": "ok"
}
```

---

### 4. POST `/api/v1/orders/:id/cancel` 用户取消订单

- 鉴权：需要登录（仅本人）。
- 路径参数：
  - `id`：订单 ID。
- 请求体（可选）：

```json
{
  "reason": "用户主动取消"
}
```

- 行为说明：
  - 仅允许在“待付款”状态取消（由 service 层校验）。
  - 取消成功后，订单状态置为已取消，未发货库存会回补（由 service 处理）。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 5. POST `/api/v1/orders/:id/pay` 模拟支付

- 鉴权：需要登录（仅本人）。
- 路径参数：
  - `id`：订单 ID。

- 行为说明：
  - 将订单标记为已支付，更新状态（由 `MarkPaid` 实现）。
  - 仅对待付款订单有效。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 6. POST `/api/v1/orders/:id/receive` 用户确认收货/完成

- 鉴权：需要登录（仅本人）。
- 路径参数：
  - `id`：订单 ID。

- 行为说明：
  - 配送单：状态为“配送中(3)”时可确认收货 → 已完成(4)。
  - 自取单：状态为“已付款(2)”时可确认完成 → 已完成(4)。
  - 仅支持本人订单。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

## 二、后台/运营侧订单 API

### 1. GET `/api/v1/admin/orders` 管理端订单列表

- 鉴权：需要登录 + 角色 `admin`。
- 查询参数：
  - `page`：页码，默认 1。
  - `limit`：每页数量，默认 20。
  - `status`：订单状态过滤，默认 0 不过滤。
  - `store_id`：门店 ID，可选。

- 响应示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": 123,
      "order_no": "202512080001",
      "user_id": 1,
      "store_id": 2,
      "total_amount": "20.00",
      "pay_amount": "20.00",
      "discount_amount": "0.00",
      "delivery_fee": "0.00",
      "status": 2,
      "pay_status": 2,
      "order_type": 1,
      "delivery_type": 1,
      "created_at": "...",
      "paid_at": "..."
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 100
}
```

---

### 2. GET `/api/v1/admin/orders/export` 管理端订单导出（CSV）

- 鉴权：admin。
- 查询参数：
  - `status`：同上。
  - `store_id`：同上。

- 行为说明：
  - 一次最多导出 10000 条。
  - 返回 `Content-Type: text/csv; charset=utf-8`，文件名 `orders_export.csv`。

- CSV 格式：

首行表头：

```text
id,order_no,store_id,user_id,pay_amount,status,pay_status,created_at
```

每行示例：

```text
123,202512080001,2,1,20.00,2,2,2025-12-08 10:00:00
```

---

### 3. GET `/api/v1/admin/orders/:id` 管理端订单详情

- 鉴权：admin。
- 路径参数：
  - `id`：订单 ID。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "order": { "...": "Order 字段" },
    "items": [ { "...": "OrderItem 字段" } ]
  },
  "message": "ok"
}
```

---

### 4. POST `/api/v1/orders/:id/deliver` 发货 / 开始配送

- 鉴权：登录 + 需要 `order:deliver` 权限。
- 路径参数：
  - `id`：订单 ID。

- 行为说明：
  - 一般用于将“已付款(2)”订单转为“配送中(3)”。
  - service 层会校验当前状态是否允许发货。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 5. POST `/api/v1/orders/:id/complete` 后台标记完成

- 鉴权：`order:complete` 权限。
- 路径参数：
  - `id`：订单 ID。

- 行为说明：
  - 用于从“配送中(3)”或特定状态直接结束为“已完成(4)”。
  - 管理侧人工确认或线下对账后操作。

- 响应示例：同上 `{ "ok": true }`。

---

### 6. POST `/api/v1/orders/:id/admin-cancel` 后台取消订单

- 鉴权：`order:cancel` 权限。
- 路径参数：
  - `id`：订单 ID。
- 请求体（可选）：

```json
{
  "reason": "后台取消原因"
}
```

- 行为说明：
  - 针对待付款订单取消，带有库存回补等逻辑。
  - handler 内部会记录取消前订单状态，便于审计。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 7. POST `/api/v1/orders/:id/refund` 后台手动退款（立即完成）

- 鉴权：`order:refund` 权限。
- 路径参数：
  - `id`：订单 ID。

- 行为说明（从 README 补充）：
  - 允许状态：已付款(2)、配送中(3)。
  - 行为：
    - 将订单置为已取消(5)，`PayStatus` 置为已退款(4)。
    - 若订单尚未发货，会回补商品/SKU/门店库存。
    - 自动回滚已使用的用户优惠券（恢复为未使用，回退使用计数）。

- 响应示例：`{ "ok": true }`。

---

### 8. POST `/api/v1/orders/:id/refund/start` 标记退款中

- 鉴权：`order:refund` 权限。
- 路径参数：
  - `id`：订单 ID。

- 行为说明：
  - 条件：`PayStatus == 已付款(2)` 且订单状态为 `已付款(2)` 或 `配送中(3)`。
  - 行为：仅将 `PayStatus` 置为 `退款中(3)`，不改变订单状态和库存。
  - 用于对接外部支付平台的异步退款场景。

- 响应示例：`{ "ok": true }`。

---

### 9. POST `/api/v1/orders/:id/refund/confirm` 确认退款完成

- 鉴权：`order:refund` 权限。
- 路径参数：
  - `id`：订单 ID。

- 行为说明：
  - 条件：`PayStatus == 退款中(3)`。
  - 行为：
    - 将订单状态置为已取消(5)，`PayStatus` 置为已退款(4)。
    - 若未发货则回补库存。
    - 回滚已使用的用户优惠券。

- 响应示例：`{ "ok": true }`。

---

## 三、订单状态与支付状态约定

- `Order.Status`：
  - 1：待付款
  - 2：已付款
  - 3：配送中
  - 4：已完成
  - 5：已取消

- `Order.PayStatus`：
  - 1：未付款
  - 2：已付款
  - 3：退款中
  - 4：已退款


