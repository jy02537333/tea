# 订单端到端流程与状态机说明

## 1. 概览

本说明文档基于当前 `tea-api` 代码与 PRD，串起「用户下单 → 支付 → 发货/自取 → 收货/完成 → 退款/取消」的完整流程，并明确各 API 对订单状态 `status` 与支付状态 `pay_status` 的影响，方便产品、前后端和测试统一认知。

---

## 2. 状态与关键字段

### 2.1 订单状态 `Order.Status`

- `1`：待付款
- `2`：已付款
- `3`：配送中
- `4`：已完成
- `5`：已取消

### 2.2 支付状态 `Order.PayStatus`

- `1`：未付款
- `2`：已付款
- `3`：退款中
- `4`：已退款

### 2.3 其他关键字段

- `order_type`：订单类型
  - `1` 商城
  - `2` 堂食
  - `3` 外卖
- `delivery_type`：配送方式
  - `1` 自取
  - `2` 配送
- `store_id`：门店 ID（0 表示非门店订单 / 平台维度）。
- `address_info`：收货信息，字符串形式存储为 JSON。
- `user_coupon_id`（在创建订单请求中）：用户优惠券 ID。

---

## 3. 端到端业务流程

### 3.1 用户侧主流程（正向链路）

1. **加入购物车**（略，见 `cart` 模块文档）
2. **从购物车创建订单**：`POST /api/v1/orders/from-cart`
   - 校验商品/SKU 上架状态与库存。
   - 根据 `delivery_type` / `order_type` / `store_id` / 优惠券等计算应付金额。
   - 创建订单与订单明细，减库存。
   - 初始状态：`status = 待付款(1)`，`pay_status = 未付款(1)`。
3. **用户支付（模拟）**：`POST /api/v1/orders/:id/pay`
   - 将订单标记为已付款：`status = 已付款(2)`，`pay_status = 已付款(2)`。
4. **商家发货 / 开始配送**：`POST /api/v1/orders/:id/deliver`（后台权限 `order:deliver`）
   - 将订单状态从 `已付款(2)` → `配送中(3)`。
5. **用户确认收货 / 完成**：`POST /api/v1/orders/:id/receive`
   - 若 `delivery_type = 配送(2)`：在 `status = 配送中(3)` 时可确认，变为 `已完成(4)`。
   - 若 `delivery_type = 自取(1)`：在 `status = 已付款(2)` 时可确认，变为 `已完成(4)`。

### 3.2 用户侧取消链路

1. **待付款订单用户取消**：`POST /api/v1/orders/:id/cancel`
   - 仅允许当前登录用户操作，且只能在 `status = 待付款(1)` 时成功。
   - 行为：
     - 订单状态置为 `已取消(5)`。
     - 回补库存（商品 / SKU / 门店库存）。
     - 如使用了优惠券，则根据实现回滚到未使用状态。

### 3.3 后台侧干预与退款链路

1. **后台取消订单**：`POST /api/v1/orders/:id/admin-cancel`
   - 权限：`order:cancel`。
   - 场景：如用户未支付，后台直接关闭订单；或纠错场景。
   - 行为：
     - 将订单置为 `已取消(5)`（通常针对 `待付款` 订单）。
     - 回补库存（若已扣减但未发货）。
2. **后台手动退款（一步到位）**：`POST /api/v1/orders/:id/refund`
   - 权限：`order:refund`。
   - 允许状态：`status = 已付款(2)` 或 `配送中(3)`，且 `pay_status = 已付款(2)`。
   - 行为：
     - 将订单状态置为 `已取消(5)`。
     - `pay_status = 已退款(4)`。
     - 若未发货：回补库存。
     - 回滚已使用优惠券（恢复可用、回退使用计数）。
3. **后台退款分步处理**：
   - **标记退款中**：`POST /api/v1/orders/:id/refund/start`
     - 条件：`pay_status = 已付款(2)` 且 `status` 为 `已付款(2)` 或 `配送中(3)`。
     - 行为：仅把 `pay_status` 改为 `退款中(3)`，不改变 `status` 与库存。
   - **确认退款完成**：`POST /api/v1/orders/:id/refund/confirm`
     - 条件：`pay_status = 退款中(3)`。
     - 行为：
       - `status = 已取消(5)`；
       - `pay_status = 已退款(4)`；
       - 未发货场景下回补库存；
       - 回滚优惠券使用记录。

---

## 4. 状态机视图

### 4.1 核心状态转移（用户视角 + 后台视角）

以 `status` 为主轴，结合 `pay_status`：

- **创建订单**（from cart）：
  - 初始：`status = 1(待付款)`，`pay_status = 1(未付款)`。

- **用户支付（/orders/:id/pay）**：
  - 前置：`status = 1`，`pay_status = 1`。
  - 后置：`status = 2(已付款)`，`pay_status = 2(已付款)`。

- **商家发货（/orders/:id/deliver）**：
  - 前置：`status = 2(已付款)`，`pay_status = 2(已付款)`。
  - 后置：`status = 3(配送中)`，`pay_status = 2(已付款)`。

- **用户确认收货（/orders/:id/receive）**：
  - 配送单：前置 `status = 3`，后置 `status = 4(已完成)`，`pay_status` 保持 `2`。
  - 自取单：前置 `status = 2`，后置 `status = 4(已完成)`，`pay_status` 保持 `2`。

- **用户取消（/orders/:id/cancel）**：
  - 前置：`status = 1(待付款)`，`pay_status = 1`。
  - 后置：`status = 5(已取消)`，`pay_status` 通常保持 `1`（未付款）。

- **后台取消（/orders/:id/admin-cancel）**：
  - 通常针对待付款单：从 `1` → `5`，`pay_status` 不变（未付款）。

- **后台直接退款（/orders/:id/refund）**：
  - 前置：`status ∈ {2,3}`，`pay_status = 2`。
  - 后置：`status = 5(已取消)`，`pay_status = 4(已退款)`。

- **后台退款分步（start → confirm）**：
  - `refund/start`：`pay_status 2 → 3`，`status` 不变；
  - `refund/confirm`：`status → 5`，`pay_status 3 → 4`。

### 4.2 自取 vs 配送的差异点

- **自取（delivery_type = 1）**：
  - 无“配送中”状态，典型路径：待付款(1) → 已付款(2) → 用户确认完成(4)。
  - 商家可以选择仍然调用 `deliver` 表示“已取餐/已出单”，但从业务上通常不需要。

- **配送（delivery_type = 2）**：
  - 典型路径：待付款(1) → 已付款(2) → 配送中(3) → 已完成(4)。
  - 用户只能在配送中(3) 状态下通过 `/receive` 完成订单。

---

## 5. 与优惠券、库存的联动

### 5.1 库存

- 创建订单：下单时减库存（商品 / SKU / 门店库存）。
- 取消 / 退款时的回补逻辑：
  - 用户取消（待付款）：一般不会扣库存或会直接回补，当前实现中在取消时确保库存回到初始状态（详见 `OrderService.CancelOrder` 实现）。
  - 后台取消 / 退款：
    - 若订单尚未发货，会回补库存；
    - 若已发货，根据实现可选择不回补库存（当前实现以测试用例为准）。

### 5.2 优惠券

- 下单使用：请求体中的 `user_coupon_id` 在 service 层会被校验并标记为已使用。
- 退款 / 取消：
  - 后台退款（包括 `refund` 和 `refund/confirm`）会自动回滚优惠券：
    - 恢复为未使用状态；
    - 回退使用计数（如有）。

---

## 6. 与 PRD 的对应关系（简要）

- 小程序端：
  - 「我的订单」列表 / 详情：对应 `GET /orders` 和 `GET /orders/:id`。
  - 「取消订单」「确认收货」入口：对应 `POST /orders/:id/cancel` 与 `POST /orders/:id/receive`。
- 后台平台：
  - 订单列表、导出、详情：`GET /admin/orders`、`GET /admin/orders/export`、`GET /admin/orders/:id`。
  - 订单运营操作：发货、完成、后台取消：`POST /orders/:id/deliver` / `complete` / `admin-cancel`。
  - 财务/售后操作：退款全流程：`POST /orders/:id/refund` / `refund/start` / `refund/confirm`。

当你在设计或调整 PRD 时，可以直接参考本状态机及 API 列表，保证文案中的“可操作按钮/流程”与后端能力一一对齐。