# 茶心阁小程序 API 服务

基于 Go + Gin + GORM + MySQL + Redis 构建的茶叶店小程序后端API服务。

## 功能特性

- 🔐 微信小程序登录认证
- 👤 用户管理系统
- 🛍️ 商品管理系统
- 📦 订单管理系统
- 💰 支付系统（微信支付、支付宝）
- 🚚 外卖平台集成（美团、饿了么、百度）
- 🎫 营销活动系统
- 📊 数据统计分析
- 📝 操作日志记录
- 🔄 提现管理系统
 - 💹 资金计息（按日复利、调度、导出、权限）

## 技术栈

- **后端框架**: Go 1.21 + Gin
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0
- **ORM**: GORM v2
- **认证**: JWT
- **日志**: Zap
- **配置管理**: Viper

## 项目结构

```
tea-api/
├── cmd/                    # 应用入口
├── internal/              # 内部应用代码
│   ├── config/           # 配置管理
│   ├── handler/          # 请求处理器
│   ├── service/          # 业务逻辑
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/       # 中间件
│   └── router/           # 路由配置
├── pkg/                   # 公共库
│   ├── database/         # 数据库连接
│   └── utils/            # 工具函数
├── configs/              # 配置文件
├── scripts/              # 脚本文件
├── docs/                 # 文档
├── go.mod
└── go.sum
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 7.0+

### 2. 克隆项目

```bash
git clone <repository-url>
cd tea-api
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 配置数据库

1. 创建 MySQL 数据库：
```sql
CREATE DATABASE tea_shop CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 执行初始化脚本：
```bash
mysql -u root -p tea_shop < scripts/init.sql
```

3. 修改配置文件 `configs/config.yaml`，设置正确的数据库连接信息。

### 5. 启动服务

#### Windows 用户
```bash
scripts\start.bat
```

#### Linux/Mac 用户
```bash
go run ./cmd -config=configs/config.yaml
```

服务将在 `:8080` 端口启动。

## 文档索引

- RBAC 权限与缓存说明：`doc/rbac.md`
- 计息功能说明：`doc/accrual.md`

## API 文档

### 基础信息
- Base URL: `http://localhost:8080/api/v1`
- 认证方式: Bearer Token (JWT)

### 用户相关接口

#### 用户登录
```http
POST /user/login
Content-Type: application/json

{
    "code": "微信登录code"
}
```

#### 获取用户信息
```http
GET /user/info
Authorization: Bearer <token>
```

#### 更新用户信息
```http
PUT /user/info
Authorization: Bearer <token>
Content-Type: application/json

{
    "nickname": "新昵称",
    "avatar": "头像URL"
}
```

### 健康检查
```http
GET /health
```

### 购物车（需登录）
- GET /cart 获取购物车条目列表
- POST /cart/items 添加条目
    - 参数: { "product_id": number, "sku_id"?: number, "quantity": number>0 }
- PUT /cart/items/:id 更新条目数量（<=0 等同删除）
- DELETE /cart/items/:id 删除条目
- DELETE /cart/clear 清空购物车

说明：同一商品+SKU 会自动合并数量；创建时会校验商品/SKU是否存在且已上架。

### 订单（需登录）
- ### 门店（部分）
- GET /stores 门店列表（支持 ?status=1&page=1&limit=20&lat=..&lng=..，返回 distance_km）
- GET /stores/:id 门店详情
- POST /stores 创建门店（需登录）
- PUT /stores/:id 更新门店（需登录）
- DELETE /stores/:id 删除门店（需登录）

- POST /orders/from-cart 从购物车下单
    - Body: { "delivery_type": 1|2, "address_info"?: string, "remark"?: string, "user_coupon_id"?: number, "store_id"?: number, "order_type"?: 1|2|3 }
    - Response: { "id": number, "order_no": string, "pay_amount": number, "discount_amount": number }
- GET /orders 订单列表（支持 ?status=1|2|...&page=1&limit=20&store_id?=N）
- GET /orders/:id 订单详情（包含 items）
- POST /orders/:id/cancel 取消订单（仅待付款可取消）
- POST /orders/:id/pay 模拟支付（仅待付款可支付）
- POST /orders/:id/receive 用户确认收货/完成订单（登录用户）
    - 配送单：状态为 配送中(3) 时可确认收货
    - 自取单：状态为 已付款(2) 时可确认完成
    - 对应管理端：
        - POST /orders/:id/deliver 发货（需权限 order:deliver，状态从已付款(2) -> 配送中(3)）
        - POST /orders/:id/complete 完成（需权限 order:complete，配送中(3) -> 已完成(4)）
        - POST /orders/:id/admin-cancel 后台取消（需权限 order:cancel，仅待付款可取消，含库存回补）
        - POST /orders/:id/refund 手动退款（需权限 order:refund，仅已付款可退款）
            - 允许状态：已付款(2)、配送中(3)
            - 若订单仍为已付款(2)：会回补商品/SKU/门店库存
            - 订单将置为 已取消(5) 且支付状态置为 已退款(4)
            - 将自动回滚已使用的用户优惠券（置回未使用并回退券 used_count）
        - POST /orders/:id/refund/start 标记退款中（需权限 order:refund）
            - 条件：PayStatus=已付款(2) 且订单状态为 已付款(2) 或 配送中(3)
            - 行为：仅将 PayStatus 置为 退款中(3)，不变更订单状态与库存
        - POST /orders/:id/refund/confirm 确认退款完成（需权限 order:refund）
            - 条件：PayStatus=退款中(3)
            - 行为：将订单置为 已取消(5)，PayStatus=已退款(4)；若未发货则回补库存；自动回滚已使用的用户优惠券

规则说明：
- 下单会校验商品/SKU是否上架，且库存足够；采用乐观扣减（库存>=数量）更新。
- 可选使用用户优惠券：支持满减、折扣、免单，校验有效期与门槛，按券类型计算优惠后生成 `discount_amount` 与 `pay_amount`（以数字返回）。
- 可选绑定门店：传 `store_id` 时校验门店启用；`order_type` 支持 1商城 2堂食 3外卖（默认1）。若门店已为该商品配置覆盖价，则按覆盖价计算行项目与订单金额；同时扣减门店维度库存。
- 取消订单会自动回补库存（商品、SKU，且若订单绑定门店则回补门店库存）。
- 用户确认收货接口仅允许订单本人调用；支持自取单在已付款状态直接确认完成、配送单在配送中状态确认收货。
- 后台取消接口用于运营介入，仅在待付款时可执行，逻辑同用户取消并回补库存。

### 退款记录（管理端）

- GET /admin/refunds 退款记录列表（需权限 order:refund）
    - 支持查询：`order_id`、`payment_id`、`refund_no`(模糊)、`status`、`start`、`end`、`page`、`limit`
    - 返回字段包含退款记录及其关联 `Order`、`Payment`
- GET /admin/refunds/export 导出退款记录（需权限 order:refund）
    - 支持 `?format=csv|xlsx`（默认 csv），同步支持上述查询条件，最多导出最近 5000 条
    - 字段：`id, refund_no, order_id, order_no, payment_id, payment_no, refund_amount, refund_reason, status, refunded_at, created_at`

### 门店库存绑定（管理端）

- GET /admin/stores/:id/products 列出门店已绑定的商品库存与覆盖价
- POST /admin/stores/:id/products 绑定/更新门店商品库存与覆盖价
    - Body: { "product_id": number, "stock": number, "price_override"?: string 数字字符串，如 "8.50" }
- DELETE /admin/stores/:id/products/:product_id 解绑门店商品
- GET /admin/stores/:id/orders/stats 门店订单统计（需管理权限）

说明：
- 门店与商品的绑定记录模型为 `StoreProduct(store_id, product_id, stock, price_override)`，(store_id, product_id) 唯一。
- `price_override` 留空或 "0" 表示不覆盖，使用商品原价。
- 订单从购物车创建时，如包含 `store_id`，系统将优先使用覆盖价并扣减门店库存；取消订单会回补门店库存。
- 列表接口与统计接口需要管理员令牌，建议通过开发登录获取 JWT 后在管理端页面调用。

### 门店维度商品列表（前台）

- GET /products?store_id=N&page=1&limit=20
    - 在常规商品字段基础上，额外返回：
        - `store_stock`: 该门店维度库存（无绑定为 null）
        - `store_price_override`: 覆盖价（字符串，未设置为 null）
    - 用途：小程序可按选中门店展示覆盖价与“门店缺货”状态。

### 操作日志（管理端）

- GET `/api/v1/admin/logs/operations` 列表，GET `/api/v1/admin/logs/operations/export` 导出 CSV/XLSX（`?format=csv|xlsx`）
    - 筛选参数：
        - `module`: 模块，例如 `finance`
        - `method`: 前缀匹配（历史兼容）
        - `operation`: 精确匹配操作名，例如 `order.refund`、`order.refund_start`、`order.refund_confirm`、`order.admin_cancel`
        - `order_id`: 按 `request_data` 中的订单ID过滤（内部使用 LIKE 匹配）
        - `user_id`: 操作者ID
        - `start`、`end`: 创建时间范围
        - `page`、`limit`: 分页（导出忽略分页，最多 5000 条）

### 支付记录（管理端）

- GET `/api/v1/admin/payments` 列表，GET `/api/v1/admin/payments/export` 导出 CSV/XLSX（`?format=csv|xlsx`）
    - 筛选参数：
        - `order_id`: 订单ID
        - `store_id`: 门店ID
        - `payment_no`: 支付编号（模糊匹配）
        - `status`: 支付状态 1待支付 2成功 3失败
        - `method`: 支付方式 1微信 2支付宝
        - `start`、`end`: 创建时间范围
        - `page`、`limit`: 分页（导出忽略分页，最多 5000 条）
    - 导出字段：`id, payment_no, order_id, order_no, store_id, store_name, method, amount, status, paid_at, created_at`

### 提现记录（管理端）

- GET `/api/v1/admin/withdraws` 列表，GET `/api/v1/admin/withdraws/export` 导出 CSV/XLSX（`?format=csv|xlsx`）
    - 筛选参数：
        - `user_id`: 用户ID
        - `withdraw_no`: 提现单号（模糊匹配）
        - `status`: 提现状态 1申请中 2处理中 3已完成 4已拒绝
        - `start`、`end`: 创建时间范围
        - `page`、`limit`: 分页（导出忽略分页，最多 5000 条）
    - 导出字段：`id, withdraw_no, user_id, amount, fee, actual_amount, type, status, processed_at, processed_by, created_at`
    - 状态操作（需权限 `order:refund`）：
        - POST `/api/v1/admin/withdraws/:id/approve` 受理（状态置为处理中2）
        - POST `/api/v1/admin/withdraws/:id/complete` 完成（状态置为已完成3）
        - POST `/api/v1/admin/withdraws/:id/reject` 拒绝（状态置为已拒绝4）
        - 请求体：`{ "remark": "可选备注/原因" }`

### 财务对账（管理端）

- GET `/api/v1/admin/finance/summary` 对账概要
    - 筛选参数：
        - `start`、`end`: 时间范围（按创建时间）
        - `store_id`：按门店过滤（可选）
        - `method`：按支付方式过滤（1微信 2支付宝，可选）
        - `group`: 可选，`day|store|method` 返回对应维度的明细 `rows`
    - 返回：
        - `summary`: `total_payments_count/amount`、`total_refunds_count/amount`、`net_amount`
        - 当 `group=day` 时，返回 `rows`：`date,pay_count,pay_amount,refund_count,refund_amount,net_amount`
        - 当 `group=store` 时，返回 `rows`：`store_id,store_name,pay_count,pay_amount,refund_count,refund_amount,net_amount`
        - 当 `group=method` 时，返回 `rows`：`method,pay_count,pay_amount,refund_count,refund_amount,net_amount`
- GET `/api/v1/admin/finance/summary/export?format=csv|xlsx&group=day|store|method` 导出汇总
    - 支持 `day|store|method` 三种维度，导出对应明细；`group=store` 导出包含 `Store Name`

- GET `/api/v1/admin/finance/reconcile/diff` 支付对账差异
    - 含义：对比订单应付金额(`orders.pay_amount`)与成功支付合计(`SUM(payments.amount)`)，返回非零差异
    - 筛选参数：
        - `start`、`end`: 时间范围（按支付创建时间）
        - `store_id`：按门店过滤（可选）
        - `method`：按支付方式过滤（可选）
        - `page`、`limit`：分页（默认20，最大200）
    - 返回 `rows` 字段：`order_id,order_no,store_id,order_pay_amount,paid_amount_sum,diff_amount`
- GET `/api/v1/admin/finance/reconcile/diff/export?format=csv|xlsx` 导出差异结果
    - 与查询相同的筛选参数，导出所有差异行

## 配置说明

配置文件位于 `configs/config.yaml`，主要配置项包括：

- `server`: 服务器配置（端口、模式等）
- `database`: MySQL 数据库配置
- `redis`: Redis 配置
- `jwt`: JWT 认证配置
- `wechat`: 微信小程序配置
- `alipay`: 支付宝配置
- `delivery`: 外卖平台配置
- `finance.accrual`: 计息与调度配置（示例见 `configs/config.yaml`）
    - `enabled`: 是否启用调度
    - `time`: 每日执行时间（HH:MM）
    - `rate`: 默认日利率（如 0.001）
    - `timezone`: 时区（如 Asia/Shanghai）
    - `skip_weekends`: 跳过周末
    - `holidays`: 节假日白名单（YYYY-MM-DD 数组）
    - `use_redis_lock`: 是否使用 Redis 分布式锁
    - `lock_ttl_second`: 锁超时时长（秒）
    - `allowed_roles`: 具备计息操作权限的角色（除 admin 外的白名单，DB 权限优先生效）

### 计息能力速览

- 用户级利率覆盖：`User.InterestRate>0` 时覆盖默认利率
- 并发与幂等：`InterestRecord` 具有 `(user_id, date)` 复合唯一索引；服务端在并发冲突时自动忽略重复
- 导出：支持 CSV/XLSX、表头中英、字段选择、可选 ZIP 打包
- 路由与权限：
    - 仅 admin：`GET /api/v1/admin/users`
    - 权限控制：
        - `POST /api/v1/admin/accrual/run` 需要 `accrual:run`
        - `GET /api/v1/admin/accrual/export` 需要 `accrual:export`
        - `GET /api/v1/admin/accrual/summary` 需要 `accrual:summary`
    - 鉴权顺序：admin → DB(用户-角色-权限) → 配置 `allowed_roles` 回退

示例：

1) 手动计提

POST /api/v1/admin/accrual/run
{
    "date": "2025-11-12",
    "rate": 0.001
}

2) 导出英文 XLSX，仅导出部分字段并打包 zip

GET /api/v1/admin/accrual/export?start=2025-11-01&end=2025-11-12&format=xlsx&lang=en&fields=user_id,date,interest_amount&zip=1

更多详情参见 `doc/accrual.md`。

### 优惠券（简化演示）
- GET /coupons 列表（支持 ?status=1）
- POST /coupons 创建（需登录；仅用于演示）
- POST /coupons/grant 发券给用户（需登录；仅用于演示）
- GET /user/coupons 当前用户可用券（需登录）

## 数据库设计

项目采用统一的审计字段设计：
- `id`: 主键ID
- `uid`: 全局唯一标识
- `created_at`: 创建时间
- `created_by`: 创建人ID
- `updated_at`: 更新时间
- `updated_by`: 更新人ID
- `deleted_at`: 软删除时间
- `is_deleted`: 删除标记

主要数据表：
- `users`: 用户表
- `roles`: 角色表
- `permissions`: 权限表
- `categories`: 商品分类表
- `products`: 商品表
- `orders`: 订单表
- `payments`: 支付记录表
- `delivery_orders`: 配送订单表

## 开发说明

### 添加新接口

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/service/` 中实现业务逻辑
3. 在 `internal/handler/` 中实现请求处理
4. 在 `internal/router/` 中注册路由

### 中间件

- `AuthMiddleware`: JWT 认证中间件
- `CORSMiddleware`: 跨域处理中间件
- `DetailedAccessLogMiddleware`: 访问日志中间件
 - `OperationLogMiddleware`: 管理端变更操作日志（可配置开关 / 白名单 / 黑名单）

### 日志接口（需 rbac:view）

- 操作日志：
    - GET `/api/v1/admin/logs/operations` 列表（过滤：module、method、path、user_id、start、end）
    - GET `/api/v1/admin/logs/operations/export` 导出 CSV/XLSX（`?format=csv|xlsx`）
- 访问日志：
    - GET `/api/v1/admin/logs/access` 列表（过滤：method、path、status、user_id、start、end）
    - GET `/api/v1/admin/logs/access/export` 导出 CSV/XLSX（`?format=csv|xlsx`）

### 配置（操作日志）

`configs/config.yaml`

```
observability:
    operationlog:
        enabled: true
        include_prefixes: ["/api/v1/admin/rbac", "/api/v1/admin/accrual"]
        exclude_prefixes: ["/api/v1/admin/logs"]
```

## CI/E2E 一键验证

- 运行单元/集成测试并执行 RBAC 自动失效演示：
    - PowerShell（Windows）：`scripts/ci_e2e.ps1`
    - 内含：`go test ./...` 与 `scripts/e2e_rbac_auto_invalidate.ps1 -StartServer`

### 工具函数

- `utils.GenerateUID()`: 生成唯一ID
- `utils.GenerateOrderNo()`: 生成订单号
- `utils.GenerateToken()`: 生成JWT token
- `utils.ParseToken()`: 解析JWT token

## 部署说明

### 生产环境配置

1. 修改 `configs/config.yaml` 中的配置：
   - 设置 `server.mode` 为 `release`
   - 配置生产数据库连接
   - 设置强密码和密钥

2. 编译生产版本：
```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tea-api ./cmd
```

3. 使用进程管理工具（如 systemd、supervisor）管理服务。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！