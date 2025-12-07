# 前后端联调 Checklist / 指南

本文为快速联调参考，基于当前本地验证结果（见 `验证结果.md`）。将这些步骤作为 QA / 前端 / 后端联调的操作指南，并把关键命令纳入 CI 验证流程。

**环境配置（标准现状）**

- MySQL: `127.0.0.1:3308`，数据库 `tea_shop`，用户 `root`，密码 `gs963852`
  - 连接字符串：`root:gs963852@tcp(127.0.0.1:3308)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local`
- Redis: `127.0.0.1:6379`，密码 `123456`
- RabbitMQ: `127.0.0.1:5672`，guest/guest

> 说明：以上配置已写入 `tea-api/configs/config.mysql.local.yaml`（用于本地重启后端服务）。

---

## 本地准备（快速检查点）

- 确认后端在本地可访问（默认端口 `9292`）：

```bash
# 在项目根目录
# 启动后端（示例，按你本地启动方式替换）
cd tea-api
# 如果使用 go run
go run . --config configs/config.mysql.local.yaml
# 或者使用已有的启动脚本 / systemd / docker-compose
```

- 确保已运行必要的 seeder（示例：创建 SKU）

```bash
cd tea-api
go run scripts/seed_skus.go -product 1 -sku_name "默认规格" -sku_code "SKU-P1-001" -price "99.00" -stock 20
```

- 前端默认 baseURL 覆盖：
  - `wx-fe` 默认基址: `http://127.0.0.1:9292`，可通过 `.env` 覆盖 `VITE_API_BASE_URL`
  - `admin-fe` 默认基址: `http://localhost:9292`，可通过 `.env` 覆盖 `VITE_API_BASE_URL`

---

## 验证脚本与常用命令

- 运行仓库内的自动验证脚本（会把每个响应存到 `build-ci-logs/api_validation/`）：

```bash
# 从仓库根目录
sh scripts/run_api_validation.sh
# 或者（如果想查看摘要）
cat build-ci-logs/api_validation/summary.txt
```

- 获取 admin dev-login token（示例）：

```bash
# 请将 openid 替换为你的 admin_openid（示例为 admin_openid）
curl -X POST 'http://127.0.0.1:9292/api/v1/user/dev-login' \
  -H 'Content-Type: application/json' \
  -d '{"openid":"admin_openid"}' \
  | jq . > build-ci-logs/admin_login_response.json

# 提取 token（示例字段名请根据实际响应调整）
cat build-ci-logs/admin_login_response.json | jq -r '.data.token'
```

- 常用接口取样（无 token）：

```bash
# 产品列表
curl 'http://127.0.0.1:9292/api/v1/products'

# 门店列表
curl 'http://127.0.0.1:9292/api/v1/stores'
```

- 添加购物车（需要用户 token）：

```bash
TOKEN=<user_token_or_admin_token>
curl -X POST 'http://127.0.0.1:9292/api/v1/cart/items' \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1, "sku_id":1, "quantity":2}'
```

- 导出 admin 订单 CSV（示例）：

```bash
TOKEN=<admin_token>
curl -G 'http://127.0.0.1:9292/api/v1/admin/orders/export' \
  -H "Authorization: Bearer $TOKEN" \
  --data-urlencode 'status=paid' \
  --output orders_export.csv
```

---

## 常见坑与建议修复

- sku_id 行为
  - 问题：某些早期请求未传 `sku_id`，后端以前可能把 `0` 写入数据库，从而触发外键约束失败。
  - 当前状态：已把 `CartItem.SkuID` 改为可空，并在保存时写 `NULL`（本地验证通过）。
  - 建议：前端在调用 `POST /api/v1/cart/items` 时优先传入有效的 `sku_id`；若业务允许没有 sku，则传空或 omit 字段。

- 缺少种子数据
  - 问题：`GET /api/v1/products/:id` 会返回 404，或 `cart` 报 500，因为相关 users/products/skus 未种子。
  - 建议：运行后端的 seeder（见上文 SKU seeder），或直接向 API 创建测试数据；在 CI 中加入一个简单的 seed 步骤以保证可重复性。

- token / RBAC
  - 获取 admin 权限（dev-login）后会得到带 `role=admin` 的 token，存储路径：`build-ci-logs/admin_login_response.json`。
  - 建议：把这个 token 注入前端（临时方式）或在请求时加上 `Authorization` 头以便调试受限接口（如 `/api/v1/admin/*`）。

- 日期 / 导出范围
  - 在导出时，请确保日期范围参数为 `YYYY-MM-DD` 格式（服务端期望），并把相同参数用于列表查询以保证导出结果一致。

---

## 文件与输出位置

- 自动化验证输出：`build-ci-logs/api_validation/`（每个请求的 body 以及 `summary.txt`）
- admin dev-login 响应：`build-ci-logs/admin_login_response.json`
- 本地临时响应（示例）：`/tmp/cart_no_sku.json`、`/tmp/cart_with_sku.json`

---

## 把验证纳入 CI 的建议

- 在 CI job 中加入步骤：
  1. 使用预置数据库镜像或测试 DB 恢复一个已知快照
  2. 启动 `tea-api`（使用 `configs/config.mysql.local.yaml` 或 CI 专用配置）
  3. 运行 seeder（创建必需的 category/product/store/sku/user）
  4. 运行 `sh scripts/run_api_validation.sh` 并保存 `build-ci-logs/api_validation/` 到 artifacts
  5. 对 `summary.txt` 做断言（例如：关键 admin GET 接口返回 200）

- 输出格式：脚本已写出各响应 body，CI 可以将这些 JSON 上传为 artifacts 供 QA 下载；或者用一个简单的脚本把它们汇总为单页 Markdown 报告。

---

## 快速排查步骤（当接口失败时）

1. 确认 `VITE_API_BASE_URL` 是否指向 `127.0.0.1:9292` 或 `localhost:9292`（根据前端运行位置调整）。
2. 检查 `build-ci-logs/admin_login_response.json` 是否存在并包含 token；若不存在，重新执行 `dev-login`。
3. 检查是否已为所需实体运行 seeder（product/store/sku/user）。
4. 查看 `build-ci-logs/api_validation/` 下对应请求的 body/response，快速定位错误码与响应体。

---

如需我把本文件与 `验证结果.md` 的摘录合并为一页给 QA 的短版报告（或生成 CI job 示例），我可以继续把一个 GitHub Actions / GitLab CI 的 job 模板加到 `docs/ci/` 下。