# Dev Login & Long-Lived Admin Token

This note explains how frontend teammates (and anyone running the CLI E2E flow) can obtain a reusable admin token in local or dev environments.

## 1. Start the API locally

The helper script builds the no-migrate binary, starts the API, and performs an admin login automatically:

```bash
# from the repo root
bash run-tea-api.sh
```

The script waits for `http://127.0.0.1:9292/api/v1/health`, then tries two login paths:

1. `POST /api/v1/auth/login` with `{"username":"admin","password":"pass"}`.
2. If the password flow fails, it calls `POST /api/v1/user/dev-login` with `{"openid":"admin_openid"}`.

Successful responses are saved under `build-ci-logs/`:

- `build-ci-logs/admin_login_response.json`
- `build-ci-logs/admin_users_response.json`

## 2. Export the token for API calls / FE proxies

Parse the login response and export a shell variable that any CLI, HTTP client, or frontend dev server can re-use:

```bash
export TOKEN=$(python3 - <<'PY'
import json
from pathlib import Path
p = Path('build-ci-logs/admin_login_response.json')
data = json.loads(p.read_text())
if isinstance(data, dict):
    data = data.get('data') or data
print(data.get('token', ''))
PY
)
export BASE_URL="http://127.0.0.1:9292/api/v1"
```

Once `TOKEN` is in your environment:

```bash
curl -H "Authorization: Bearer $TOKEN" "$BASE_URL/admin/users?page=1&limit=20"
```

The CLI E2E (`go run ./cmd/e2e_single_sku_order`) also honors `TOKEN`. If the variable is empty it will call `/user/dev-login` on its own, so FE teams can rely on either approach.

## 3. Bootstrap command for FE teammates

To get a local backend plus a ready-to-use `TOKEN` in one shot:

```bash
bash run-tea-api.sh && \
export BASE_URL="http://127.0.0.1:9292/api/v1" && \
export TOKEN=$(python3 - <<'PY'
import json
from pathlib import Path
p = Path('build-ci-logs/admin_login_response.json')
data = json.loads(p.read_text())
if isinstance(data, dict):
    data = data.get('data') or data
print(data.get('token', ''))
PY
)
```

> Feel free to swap the inline Python with `jq` or another parser. After exporting `TOKEN` and `BASE_URL`, your frontend dev server can proxy to `$BASE_URL` and forward the bearer token automatically.

Keep this file handy when sharing onboarding steps—everybody now has the exact command sequence to spin up the backend and obtain a stable admin token.

## 4. E2E CLI（单 SKU 支付验证）

完成上面的 `run-tea-api.sh` + `TOKEN/BASE_URL` 导出后，可直接运行 `tea-api/cmd/e2e_single_sku_order` 来验证“下单 → 支付 → 回调 → 查询订单”主链路：

```bash
cd tea-api
export BASE_URL="http://127.0.0.1:9292/api/v1"
# TOKEN 可选，留空会自动调用 /user/dev-login（openid=admin_openid）
go run ./cmd/e2e_single_sku_order
```

脚本会依次打印 `/products`、`/cart`、`/orders/from-cart`、`/payments/unified-order`、`/payments/callback`、`/orders/:id` 的响应。其中回调阶段会自动带上 `test_mode=true`，只要 `configs/config.yaml` 中 `system.env` 为 `local` 或 `dev` 即可跳过签名校验。更详细的参数/报文说明，可见 `docs/payment-flow.md`。
