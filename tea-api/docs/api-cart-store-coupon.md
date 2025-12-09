# 购物车 / 门店 / 优惠券 API 文档

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

## 一、购物车 API（Cart）

路由前缀：`/cart`，所有接口均需要登录。

### 1. GET `/api/v1/cart` 列出购物车条目

- 鉴权：需要登录。
- 请求参数：无。
- 响应示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "cart_id": 1,
      "product_id": 10,
      "sku_id": 100,
      "quantity": 2,
      "created_at": "...",
      "updated_at": "...",
      "product": {
        "id": 10,
        "name": "下单商品",
        "price": "10.00",
        "status": 1
      },
      "sku": {
        "id": 100,
        "sku_name": "100g",
        "price": "10.00",
        "status": 1
      }
    }
  ],
  "message": "ok"
}
```

> 实际字段以 `CartItem` 模型及其 `Product`/`Sku` 预加载结果为准，核心字段为 `product_id` / `sku_id` / `quantity`。

---

### 2. POST `/api/v1/cart` / `/api/v1/cart/items` 添加购物车条目

- 鉴权：需要登录。
- 路径说明：`POST /api/v1/cart` 与 `POST /api/v1/cart/items` 在当前实现中等价，推荐在新代码中统一使用 `POST /api/v1/cart`，`/items` 作为兼容旧调用保留。
- 请求体：

```json
{
  "product_id": 10,
  "sku_id": 100,
  "quantity": 2
}
```

- 字段说明：
  - `product_id`：必填，商品 ID。
  - `sku_id`：可选，SKU ID；若为空则为无规格商品。
  - `quantity`：必填，数量（>0）。

- 行为说明：
  - 会校验商品和 SKU 是否存在且处于上架状态。
  - 若相同商品+SKU 已存在，`CartService.AddItem` 会按实现决定是累加数量还是合并逻辑（当前实现为更新/创建逻辑，可以在后续技术设计文档中固定）。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "cart_id": 1,
    "product_id": 10,
    "sku_id": 100,
    "quantity": 2
  },
  "message": "ok"
}
```

---

### 3. PUT `/api/v1/cart/items/:id` 更新购物车条目数量

- 鉴权：需要登录。
- 路径参数：
  - `id`：购物车条目 ID。
- 请求体：

```json
{
  "quantity": 3
}
```

- 行为说明：
  - `quantity > 0`：更新数量。
  - `quantity <= 0`：等同删除该条目。
  - 仅允许修改当前登录用户的购物车条目。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 4. DELETE `/api/v1/cart/items/:id` 删除购物车条目

- 鉴权：需要登录。
- 路径参数：
  - `id`：购物车条目 ID。

- 行为说明：
  - 删除当前用户购物车中的指定条目；非本人条目无法删除。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 5. DELETE `/api/v1/cart/clear` 清空购物车

- 鉴权：需要登录。
- 请求参数：无。
- 行为说明：
  - 删除当前用户购物车下的所有条目。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

> 说明：从购物车创建订单的接口见 `api-orders.md` 中的 `POST /orders/from-cart`，会基于当前购物车条目生成订单并清空购物车。

---

## 二、门店 API（Store）

路由前缀：`/stores`；创建、更新、删除需要登录（通常由平台/门店管理员使用），列表和详情为公开接口。

### 1. GET `/api/v1/stores` 门店列表

- 鉴权：无需登录。
- 查询参数：
  - `page`：页码，默认 `1`。
  - `limit`：每页条数，默认 `20`。
  - `status`：门店状态，可选（如 1=正常营业，其他状态参考 `Store.Status` 枚举）。
  - `lat` / `lng`：可选，用户当前位置，用于计算距离与按距离排序。

- 响应示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "西湖路旗舰店",
      "address": "杭州市西湖区...",
      "phone": "0571-12345678",
      "latitude": 30.123,
      "longitude": 120.123,
      "business_hours": "10:00-22:00",
      "images": "[\"https://...jpg\"]",
      "status": 1,
      "distance_km": 2.35
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 1
}
```

> 实际字段以 `Store` 模型为准；`distance_km` 由 `StoreService.ListStores` 根据经纬度计算并附加。

---

### 2. GET `/api/v1/stores/:id` 门店详情

- 鉴权：无需登录。
- 路径参数：
  - `id`：门店 ID。
- 响应示例：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "西湖路旗舰店",
    "address": "...",
    "phone": "...",
    "latitude": 30.123,
    "longitude": 120.123,
    "business_hours": "10:00-22:00",
    "images": "[\"https://...jpg\"]",
    "status": 1
  },
  "message": "ok"
}
```

---

### 3. POST `/api/v1/stores` 创建门店

- 鉴权：需要登录（通常仅平台/运营或具有门店管理权限的账号可调用）。
- 请求体：

```json
{
  "name": "西湖路旗舰店",
  "address": "杭州市西湖区...",
  "phone": "0571-12345678",
  "latitude": 30.123,
  "longitude": 120.123,
  "business_hours": "10:00-22:00",
  "images": "[\"https://...jpg\"]",
  "status": 1
}
```

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "西湖路旗舰店",
    "address": "...",
    "status": 1
  },
  "message": "ok"
}
```

---

### 4. PUT `/api/v1/stores/:id` 更新门店

- 鉴权：需要登录。
- 路径参数：
  - `id`：门店 ID。
- 请求体：同创建门店，但字段均为可选；未传入的字段可保持不变。

- 行为说明：
  - 更新门店的基础信息与状态，`updated_at` 会被自动更新为当前时间。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 5. DELETE `/api/v1/stores/:id` 删除门店

- 鉴权：需要登录。
- 路径参数：
  - `id`：门店 ID。

- 行为说明：
  - 通常为软删除或逻辑删除（具体取决于 `StoreService.DeleteStore` 实现）；需确保不影响历史订单数据的查询。

- 响应示例：

```json
{
  "code": 0,
  "data": { "ok": true },
  "message": "ok"
}
```

---

### 6. GET `/api/v1/admin/stores/:id/orders/stats` 门店订单统计

- 鉴权：需要登录 + 角色 `admin`（挂在 `/admin` 分组下）。
- 路径参数：
  - `id`：门店 ID。

- 行为说明：
  - 使用 `OrderService.GetStoreOrderStats` 统计指定门店的订单数据（如订单数、成交额、退款情况等），具体字段以该方法返回结构为准。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "total_orders": 100,
    "paid_orders": 80,
    "refund_orders": 5,
    "total_amount": "12345.67"
  },
  "message": "ok"
}
```

---

## 三、优惠券 API（Coupon）

优惠券相关路由包括：
- `/coupons`：平台配置与发券入口（管理视角）。
- `/user/coupons`：当前用户的可用优惠券列表。

### 1. POST `/api/v1/coupons` 创建优惠券

- 鉴权：需要登录（建议在权限体系中仅开放给运营/管理员）。
- 请求体：

```json
{
  "name": "满100减20元",
  "type": 1,
  "amount": "20.00",
  "discount": "0",
  "min_amount": "100.00",
  "total_count": 1000,
  "status": 1,
  "start_time": "2025-01-01T00:00:00Z",
  "end_time": "2025-12-31T23:59:59Z",
  "description": "新用户专享活动券"
}
```

- 字段说明：
  - `type`：优惠券类型
    - `1` 满减券：`amount` 为减免金额。
    - `2` 折扣券：`discount` 为折扣（0-1 之间的小数，如 0.9 表示 9 折）。
    - `3` 免单券：免当次订单全额。
  - `amount`：满减金额，字符串形式的 decimal。
  - `discount`：折扣系数，字符串形式的 decimal。
  - `min_amount`：使用门槛金额。
  - `total_count`：发放总量。
  - `status`：状态（1=启用，其它值为禁用/下线）。
  - `start_time` / `end_time`：生效时间范围，RFC3339 格式。
  - 校验规则（同 `CouponService.CreateCoupon`）：
    - 名称必填、类型范围必须在 1~3；
    - 满减券金额需大于 0；折扣券折扣系数在 (0,1]；
    - `total_count` > 0 且 `end_time` 晚于 `start_time`。
    - 违反上述任一约束会返回 `400` 且 `message` 为具体错误原因（例如 “满减金额需大于0”）。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "满100减20元",
    "type": 1,
    "amount": "20.00",
    "min_amount": "100.00",
    "status": 1
  },
  "message": "ok"
}
```

---

### 2. GET `/api/v1/coupons` 列出优惠券

- 鉴权：无需登录。
- 查询参数：
  - `status`：可选，按状态过滤（如 1=启用）。

- 行为说明：
  - 返回当前符合条件的优惠券定义列表，方便后台界面展示与选择。

- 响应示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "满100减20元",
      "type": 1,
      "amount": "20.00",
      "min_amount": "100.00",
      "status": 1,
      "start_time": "2025-01-01T00:00:00Z",
      "end_time": "2025-12-31T23:59:59Z"
    }
  ],
  "message": "ok"
}
```

---

### 3. POST `/api/v1/coupons/grant` 给用户发券

- 鉴权：需要登录。
- 请求体：

```json
{
  "coupon_id": 1,
  "user_id": 123
}
```

- 行为说明：
  - 若 `user_id` 传入 0 或未传，接口会尝试使用当前登录用户作为发券对象。
  - 当前实现仅检查优惠券存在且 `status=1`，尚未限制有效期或库存；请在调用侧自行控制发券策略。

- 响应示例：

```json
{
  "code": 0,
  "data": {
    "id": 100,
    "user_id": 123,
    "coupon_id": 1,
    "status": 1
  },
  "message": "ok"
}
```

---

### 4. GET `/api/v1/user/coupons` 当前用户可用优惠券

- 鉴权：需要登录。
- 请求参数：无（当前实现仅返回“可用”列表）。

- 行为说明：
  - 返回当前登录用户在有效期内、状态为“可用”的所有优惠券实例。
  - 仅返回 `user_coupons.status = 1` 且对应 `coupons.status = 1`、`当前时间` 落在 `start_time ~ end_time` 范围内的记录。
  - 订单创建时通过 `user_coupon_id` 关联到具体 `UserCoupon` 记录，创建成功后会把该券标记为已使用（`status` -> 2，写入 `used_at`、`order_id`）。

- 响应示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": 100,
      "user_id": 123,
      "coupon_id": 1,
      "status": 1,
      "coupon": {
        "id": 1,
        "name": "满100减20元",
        "type": 1,
        "amount": "20.00",
        "min_amount": "100.00",
        "start_time": "2025-01-01T00:00:00Z",
        "end_time": "2025-12-31T23:59:59Z"
      }
    }
  ],
  "message": "ok"
}
```

---

## 四、与订单模块的关系（摘要）

- 购物车 → 订单：
  - `POST /orders/from-cart` 会读取当前用户购物车条目，根据商品和门店生成订单，并在事务中清空购物车。
- 门店 → 订单：
  - 订单中的 `store_id` 字段用于标识门店订单，与 `Store` 关联；
  - `GET /admin/stores/:id/orders/stats` 用于按门店维度统计订单数据。
- 优惠券 → 订单：
  - 创建订单时通过 `user_coupon_id` 应用优惠券折扣（`user_coupon_id` 省略或置 0 表示不使用优惠券）；
  - `POST /orders/from-cart` 会按以下顺序校验优惠券：
    - 归属校验：`user_coupon.user_id` 必须等于当前用户，`status` 需为 1（兼容历史数据会兜底 0/1，>1 会返回 “无效的用户优惠券”）。
    - 有效期/状态：`coupons.status=1` 且当前时间在 `start_time~end_time` 内，否则返回 “优惠券不在有效期或已禁用”。
    - 金额门槛：购物车应付金额需 ≥ `coupon.min_amount`，否则报 “未满足优惠券使用门槛”。
    - 类型判定：
      - 满减券直接减 `amount`；
      - 折扣券按 `1-discount` 计算，`discount` 不在 (0,1] 会报 “非法的折扣券配置”；
      - 免单券折扣额等于商品总额；未知类型报 “未知的优惠券类型”。
    - 最终 `discount_amount` 不会超过 `total_amount`，`pay_amount = total_amount - discount_amount`。
  - 订单取消/退款时会自动回滚 `UserCoupon` 的使用状态（详见 `api-orders.md` 和 `order-flow-and-states.md`）。
