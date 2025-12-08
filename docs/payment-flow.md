# 支付统一下单与回调使用说明

本说明覆盖 Sprint A 新增的支付链路：`/api/v1/payments/unified-order` 与 `/api/v1/payments/callback`，并给出本地/CI 可复用的验证脚本。

## 1. 接口概览

### POST /api/v1/payments/unified-order
- **鉴权**：登录用户（Bearer Token）
- **请求体**：
  ```json
  {"order_id": 81, "method": 1}
  ```
  - `method` 默认为 1（微信），可扩展至 2（支付宝）。
- **响应示例**：
  ```json
  {
    "code": 0,
    "data": {
      "payment_no": "P20251208205144976814796",
      "order_id": 82,
      "amount": "10",
      "prepay_id": "mock_prepay",
      "nonce_str": "...",
      "timestamp": 1733652704,
      "package": "Sign=WXPay",
      "sign": "4C57...",
      "pay_url": "https://pay.example/mock"
    }
  }
  ```
- **实现要点**：
  - 若订单已有「待支付」流水，会直接复用并重新生成签名字段；避免重复扣款。
  - 签名基于 `configs/config.yaml` 下的 `wechat.api_key`，当前走 MD5(mock) 实现，可方便替换为正式 SDK。

### POST /api/v1/payments/callback
- **鉴权**：无需登录（支付渠道回调）。
- **请求体**：
  ```json
  {
    "app_id": "tea-app-mock",
    "payment_no": "P20251208205144976814796",
    "transaction_id": "mock_P20251208205144976814796",
    "trade_state": "SUCCESS",
    "paid_at": "2025-12-08T20:51:44+08:00",
    "sign": "4C57...",
    "test_mode": true
  }
  ```
- **test_mode**：当 `system.env` 为 `local` 或 `dev` 时可以传 `true` 以跳过签名校验，方便 CLI/E2E 调用；其余环境会拒绝。
- **服务端行为**：
  1. 验证签名（或 test_mode 允许的跳过）。
  2. 更新 `payments` 状态为成功并记录 `transaction_id`、`paid_at`、原始报文。
  3. 同步把 `orders.status`、`orders.pay_status` 更新为「已付款」，写入支付时间。

## 2. 自动化验证（CLI）

在 `tea-api` 目录提供 `cmd/e2e_single_sku_order`，串联了以下步骤：
1. `/products` 获取可售商品。
2. `/cart` 加购。
3. `/orders/from-cart` 下单。
4. `/payments/unified-order` 生成支付参数。
5. `/payments/callback`（`test_mode=true`）模拟渠道回调。
6. `/orders/:id` 校验状态。

### 使用方式
```bash
# 启动 tea-api（监听 9292）后执行
export BASE_URL="http://127.0.0.1:9292/api/v1"
# TOKEN 可选；留空时 CLI 会调用 /user/dev-login (openid=admin_openid)
export TOKEN="<your_admin_jwt>"

cd tea-api
go run ./cmd/e2e_single_sku_order
```
输出会打印每一步的原始响应以及 `payment_no`/`order_id`，若任一步骤失败会直接退出（非零 code）。

## 3. 常见排查

- **404: /payments/unified-order**：通常是后端进程仍是旧版本，请重新 `go build && ./tea-api` 或使用 `run-tea-api.sh`。
- **签名校验失败**：确认 `configs/config.yaml` 中的 `wechat.api_key` 与 CLI/测试使用的值一致；本地可直接开启 `test_mode`。
- **订单状态未更新**：查看 `payments` 表中 `status` 是否为 2，若否通常是回调报文未携带正确的 `payment_no`；也可在 `logs/` 中搜索 `HandleCallback`。
```}