# 扫码点餐模拟支付流程

## 功能说明

本文档介绍如何在开发环境中测试"扫码点餐→下单→模拟支付"的完整流程。

## 前置条件

1. tea-api 服务已启动（端口 9292）
2. 数据库已完成迁移（包含 orders 表的 table_id/table_no 字段）
3. 至少有一个门店（ID=1）已关联商品

## 快速测试

运行端到端测试脚本：

```bash
bash scripts/test_table_order_simple.sh
```

该脚本会自动执行以下步骤：
1. 获取测试用户 token（通过 dev-login）
2. 使用门店已关联的商品 ID
3. 添加商品到购物车
4. 创建订单（携带桌号参数 table_id/table_no）
5. 调用模拟支付接口
6. 验证订单状态和桌号信息

## 模拟支付 API

### 接口

```
POST /api/v1/orders/:id/pay
```

### 说明

- **开发环境专用**：直接将订单状态标记为已付款，不涉及真实支付渠道
- **无需支付参数**：系统自动将 `status` 设置为 2（已付款），`pay_status` 设置为 2（已支付）
- **仅本地/dev 环境可用**：生产环境禁用此接口

### 请求示例

```bash
# 假设订单 ID 为 123
curl -X POST http://127.0.0.1:9292/api/v1/orders/123/pay \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "ok": true
  }
}
```

## 桌号参数说明

下单时可携带以下参数：

- `table_id`（number）：桌号数字ID，用于内部关联
- `table_no`（string）：门店自定义桌号，如 "A12"、"大厅3号"等

示例请求体：

```json
{
  "delivery_type": 1,
  "store_id": 1,
  "order_type": 2,
  "table_id": 1,
  "table_no": "A12",
  "remark": "少冰"
}
```

订单创建后，可通过 `GET /api/v1/orders/:id` 查看桌号信息。

## 环境变量

测试脚本支持以下环境变量：

- `API_BASE`：API 地址（默认：http://127.0.0.1:9292）
- `STORE_ID`：门店 ID（默认：1）
- `TABLE_ID`：桌号 ID（默认：1）
- `TABLE_NO`：桌号编号（默认：A12）

## 注意事项

1. **数据库迁移**：首次使用前需运行 `go run -tags demo tea-api/cmd/migrate.go`
2. **服务重启**：迁移后需重新编译并重启 tea-api 服务
3. **商品准备**：确保门店至少关联了一个商品（可通过 admin-fe 或 `scripts/prepare_table_order_data.sh`）
4. **模拟支付限制**：仅限开发环境，生产环境会返回错误

## 相关文档

- [PRD.md](doc/prd.md) - 扫码进店与桌号约定详细说明
- [DEPLOY.md](DEPLOY.md) - 部署配置说明
