#!/usr/bin/env bash
set -euo pipefail

# Enhanced local API check: perform dev-login to get a token, then run user-facing API checks
# Logs and response bodies are saved under build-ci-logs/api_validation_stateful

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
BASE_URL=${BASE_URL:-"http://127.0.0.1:9292"}
OUT_DIR="$ROOT_DIR/build-ci-logs/api_validation_stateful"
mkdir -p "$OUT_DIR"

log() {
  echo "[$(date +%Y-%m-%dT%H:%M:%S%z)] $*"
}

save_body() {
  local path="$1"
  local fname="$2"
  mkdir -p "$OUT_DIR/$(dirname "$fname")"
  echo "$path" > "$OUT_DIR/$fname"
}

curl_json() {
  # args: METHOD PATH AUTH_FLAG BODY_JSON
  local method="$1"; shift
  local path="$1"; shift
  local auth="$1"; shift
  local body_json="${1:-}"; shift || true

  local url="$BASE_URL$path"
  log "=== $method $url (auth=${auth}) ==="

  if [[ -n "$body_json" ]]; then
    resp=$(curl -sS -X "$method" -H 'Content-Type: application/json' ${AUTH_HEADER:+-H "$AUTH_HEADER"} -d "$body_json" "$url" || true)
  else
    resp=$(curl -sS -X "$method" ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$url" || true)
  fi

  # Save response body
  safe_path=$(echo "$path" | sed 's#[^a-zA-Z0-9/_-]#_#g' | sed 's#/#__#g')
  out_file="${method}__${safe_path}.json"
  echo "$resp" > "$OUT_DIR/$out_file"

  # Print status summary
  top_keys=$(echo "$resp" | python3 - <<'PY'
import sys, json
try:
    obj=json.loads(sys.stdin.read())
    if isinstance(obj, dict):
        print("Top-level keys:", sorted(list(obj.keys())))
    else:
        print("Top-level keys: [non-object]")
except Exception as e:
    print("Top-level parse error:", e)
PY
  )
  echo "$top_keys"
}

# 1) Health check (no auth)
curl_json GET /api/v1/health no

# 2) Dev login to obtain token
DEV_LOGIN_BODY='{"open_id":"user_openid_local_stateful","phone":"18880000001","nickname":"StatefulUser"}'
resp=$(curl -sS -X POST -H 'Content-Type: application/json' -d "$DEV_LOGIN_BODY" "$BASE_URL/api/v1/user/dev-login")
echo "$resp" > "$OUT_DIR/POST__api_v1_user_dev-login.json"

# Extract token (if available) and set AUTH_HEADER
TOKEN=$(echo "$resp" | python3 - <<'PY'
import sys, json
try:
    data=json.loads(sys.stdin.read())
    # try common token locations
    if isinstance(data, dict):
        # format {code,message,data:{token:..}} or {token:..}
        token = None
        if 'data' in data and isinstance(data['data'], dict):
            token = data['data'].get('token')
        if not token:
            token = data.get('token')
        if token:
            print(token)
        else:
            print("")
    else:
        print("")
except Exception:
    print("")
PY
)

if [[ -n "$TOKEN" ]]; then
  AUTH_HEADER="Authorization: Bearer $TOKEN"
  log "Obtained token via dev-login."
else
  AUTH_HEADER=""
  log "Dev-login did not return a token; proceeding without Authorization header."
fi

# 3) User-facing endpoints with auth to avoid FK issues
curl_json GET /api/v1/user/info yes
curl_json GET /api/v1/cart yes
curl_json GET /api/v1/user/coupons yes

# 4) Basic catalog endpoints (no auth)
curl_json GET /api/v1/products no
curl_json GET /api/v1/categories no
curl_json GET /api/v1/stores no

# 5) Minimal success path assertions: cart add -> cart get -> order create (conditional)
if [[ -n "$AUTH_HEADER" ]]; then
  log "Attempting minimal cart/order flow with existing catalog"
  # Fetch products to pick one
  products_json=$(curl -sS -X GET ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/products?page=1&size=10" || true)
  echo "$products_json" > "$OUT_DIR/GET__api_v1_products_page_1_size_10.json"
  prod_id=$(echo "$products_json" | python3 - <<'PY'
import sys, json
try:
  data=json.loads(sys.stdin.read())
  if isinstance(data, dict):
    items = data.get('data') or []
    if isinstance(items, list) and items:
      p = items[0]
      pid = p.get('id')
      print(pid if pid is not None else "")
    else:
      print("")
  else:
    print("")
except Exception:
  print("")
PY
  )

  if [[ -n "$prod_id" ]]; then
    # Try product detail to get sku_id if present
    detail_json=$(curl -sS -X GET ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/products/$prod_id" || true)
    echo "$detail_json" > "$OUT_DIR/GET__api_v1_products_${prod_id}.json"
    # Record product detail CSV: status and sku list presence
    if [[ -f "$REPORT_FILE" ]]; then
      pstatus=$(curl -s -o /dev/null -w '%{http_code}' ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/products/$prod_id" || true)
      pok=$(python3 - <<'PY'
import sys, json
path=sys.argv[1]
try:
  d=json.loads(open(path, 'r', encoding='utf-8').read())
  sku=d.get('sku')
  print('true' if isinstance(sku, list) else 'false')
except Exception:
  print('false')
PY
      "$OUT_DIR/GET__api_v1_products_${prod_id}.json")
      ok_combined=false
      if [[ "$pstatus" == "200" && "$pok" == "true" ]]; then ok_combined=true; fi
      record "product_detail" "GET" "$BASE_URL/products/$prod_id" "$pstatus" "$ok_combined" "sku_list=$pok"
    fi
    sku_id=$(echo "$detail_json" | python3 - <<'PY'
import sys, json
try:
  d=json.loads(sys.stdin.read())
  if isinstance(d, dict):
    sku=d.get('sku') or []
    if isinstance(sku, list) and sku:
      sid=sku[0].get('id') or sku[0].get('sku_id')
      print(sid if sid is not None else "")
    else:
      print("")
  else:
    print("")
except Exception:
  print("")
PY
    )
    qty=1
    add_body=$(python3 - <<PY
import json
print(json.dumps({"product_id": int("$prod_id"), "sku_id": (int("$sku_id") if "$sku_id" else None), "quantity": $qty}))
PY
    )
    # POST /cart
    add_resp=$(curl -sS -X POST -H 'Content-Type: application/json' ${AUTH_HEADER:+-H "$AUTH_HEADER"} -d "$add_body" "$BASE_URL/api/v1/cart" || true)
    echo "$add_resp" > "$OUT_DIR/POST__api_v1_cart.json"
    # GET /cart
    cart_resp=$(curl -sS -X GET ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/cart" || true)
    echo "$cart_resp" > "$OUT_DIR/GET__api_v1_cart.json"
    # Try to create order from cart items (minimal fields)
    order_body=$(echo "$cart_resp" | python3 - <<'PY'
import sys, json
try:
  c=json.loads(sys.stdin.read())
  items=[]
  if isinstance(c, dict):
    for it in (c.get('items') or []):
      pid=it.get('product_id'); sid=it.get('sku_id'); qty=it.get('qty') or it.get('quantity') or 1
      one={"product_id": pid, "sku_id": sid, "qty": qty}
      if pid:
        items.append(one)
  body={"user_id": 1, "items": items, "delivery_type": "store", "pay_method": "wechat"}
  print(json.dumps(body))
except Exception:
  print(json.dumps({"user_id":1,"items":[],"delivery_type":"store","pay_method":"wechat"}))
PY
    )
    create_resp=$(curl -sS -X POST -H 'Content-Type: application/json' ${AUTH_HEADER:+-H "$AUTH_HEADER"} -d "$order_body" "$BASE_URL/api/v1/orders" || true)
    echo "$create_resp" > "$OUT_DIR/POST__api_v1_orders.json"
    # Brief summary
    log "Cart+Order minimal flow attempted (product=$prod_id, sku=${sku_id:-})"
  else
    log "No products returned; skipping cart/order minimal flow"
  fi
else
  log "No auth token; skipping cart/order minimal flow"
fi

# 6) Store detail menus assertion: pick first store id and check menus
stores_json=$(curl -sS -X GET ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/stores?page=1&size=10" || true)
echo "$stores_json" > "$OUT_DIR/GET__api_v1_stores_page_1_size_10.json"
store_id=$(echo "$stores_json" | python3 - <<'PY'
import sys, json
try:
  data=json.loads(sys.stdin.read())
  if isinstance(data, dict):
    items = data.get('data') or []
    if isinstance(items, list) and items:
      s = items[0]
      sid = s.get('id')
      print(sid if sid is not None else "")
    else:
      print("")
  else:
    print("")
except Exception:
  print("")
PY
)
if [[ -n "$store_id" ]]; then
  store_detail=$(curl -sS -X GET ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/stores/$store_id" || true)
  echo "$store_detail" > "$OUT_DIR/GET__api_v1_stores_${store_id}.json"
  log "Parsed store detail for id=$store_id"
  # CSV status for store detail
  status=$(curl -s -o /dev/null -w '%{http_code}' ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/stores/$store_id" || true)
  python3 - <<'PY'
import sys, json
try:
  data=json.loads(open(sys.argv[1], 'r', encoding='utf-8').read())
  menus = data.get('menus')
  if isinstance(menus, list):
    print(f"menus entries: {len(menus)}")
    # peek first category structure
    if menus:
      cat = menus[0]
      cat_keys = list(cat.keys()) if isinstance(cat, dict) else []
      print("menus[0] keys:", cat_keys)
      items = cat.get('items') if isinstance(cat, dict) else None
      if isinstance(items, list):
        print(f"menus[0].items entries: {len(items)}")
      else:
        print("menus[0].items: [not-list]")
  else:
    print("menus: [missing or non-list]")
except Exception as e:
  print("store detail parse error:", e)
PY
  "$OUT_DIR/GET__api_v1_stores_${store_id}.json"
  # Record into CSV: ok if status==200 and menus is list
  if [[ -f "$REPORT_FILE" ]]; then
    ok=$(python3 - <<'PY'
import sys, json
path=sys.argv[1]
try:
  data=json.loads(open(path, 'r', encoding='utf-8').read())
  menus = data.get('menus')
  print('true' if isinstance(menus, list) else 'false')
except Exception:
  print('false')
PY
    "$OUT_DIR/GET__api_v1_stores_${store_id}.json")
    # ok requires both status 200 and menus list
    ok_combined=false
    if [[ "$status" == "200" && "$ok" == "true" ]]; then ok_combined=true; fi
    record "store_detail" "GET" "$BASE_URL/stores/$store_id" "$status" "$ok_combined" "menus_list=$ok"
  fi
else
  log "No stores returned; skipping store detail menus assertion"
  # Record skipped in CSV with 404 to indicate no data
  if [[ -f "$REPORT_FILE" ]]; then
    record "store_detail" "GET" "$BASE_URL/stores/{id}" "404" "false" "no stores"
  fi
fi

log "Stateful API validation completed. Bodies under: $OUT_DIR"
#!/usr/bin/env bash
set -euo pipefail
BASE_URL=${BASE_URL:-http://127.0.0.1:9292/api/v1}
REPORT_DIR=${REPORT_DIR:-./build-ci-logs}
REPORT_FILE="$REPORT_DIR/local_api_check.csv"
SUMMARY_FILE="$REPORT_DIR/local_api_summary.txt"
TOKEN_FILE="$REPORT_DIR/local_api_token.txt"
mkdir -p "$REPORT_DIR"

pass_count=0
fail_count=0

echo "case,method,url,status,ok,notes" > "$REPORT_FILE"

record() {
  local case="$1"; shift
  local method="$1"; shift
  local url="$1"; shift
  local status="$1"; shift
  local ok="$1"; shift
  local notes="$*"
  echo "$case,$method,$url,$status,$ok,$notes" >> "$REPORT_FILE"
  if [[ "$ok" == "true" ]]; then
    pass_count=$((pass_count+1))
  else
    fail_count=$((fail_count+1))
  fi
}

# 1) Health
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || true)
record "health" "GET" "$BASE_URL/health" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""

# 2) Categories (public)
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/categories" || true)
record "categories_list" "GET" "$BASE_URL/categories" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""

# 3) Dev login to get token (local only)
login_json=$(curl -s -X POST "$BASE_URL/user/dev-login" -H 'Content-Type: application/json' -d '{"openid":"admin_openid"}' || true)
login_code=$(echo "$login_json" | jq -r '.code // empty' 2>/dev/null || echo "")
login_token=$(echo "$login_json" | jq -r '.data.token // empty' 2>/dev/null || echo "")
if [[ -n "$login_token" ]]; then
  echo "$login_token" > "$TOKEN_FILE"
  record "dev_login" "POST" "$BASE_URL/user/dev-login" "200" "true" "token acquired"
else
  # 尝试从data中无code结构直接读取token
  login_token=$(echo "$login_json" | jq -r '.token // empty' 2>/dev/null || echo "")
  if [[ -n "$login_token" ]]; then
    echo "$login_token" > "$TOKEN_FILE"
    record "dev_login" "POST" "$BASE_URL/user/dev-login" "200" "true" "token acquired (flat)"
  else
    record "dev_login" "POST" "$BASE_URL/user/dev-login" "${login_code:-000}" "false" "response=$login_json"
  fi
fi

AUTH_H=""
if [[ -f "$TOKEN_FILE" ]]; then
  TOKEN=$(cat "$TOKEN_FILE")
  AUTH_H="Authorization: Bearer $TOKEN"
fi

# 4) Users me (need auth)
if [[ -n "$AUTH_H" ]]; then
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/user/info" || true)
  record "user_info" "GET" "$BASE_URL/user/info" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""
else
  record "user_info" "GET" "$BASE_URL/user/info" "401" "false" "no token"
fi

# 5) Admin orders list (need auth)
if [[ -n "$AUTH_H" ]]; then
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/admin/orders?limit=5" || true)
  record "admin_orders" "GET" "$BASE_URL/admin/orders?limit=5" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""
else
  record "admin_orders" "GET" "$BASE_URL/admin/orders?limit=5" "401" "false" "no token"
fi

# 6) Tickets list (need auth)
if [[ -n "$AUTH_H" ]]; then
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/admin/tickets?limit=5" || true)
  record "admin_tickets" "GET" "$BASE_URL/admin/tickets?limit=5" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""
else
  record "admin_tickets" "GET" "$BASE_URL/admin/tickets?limit=5" "401" "false" "no token"
fi

# 7) Dashboard todos (need auth)
if [[ -n "$AUTH_H" ]]; then
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/admin/dashboard/todos" || true)
  record "admin_dashboard_todos" "GET" "$BASE_URL/admin/dashboard/todos" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""
else
  record "admin_dashboard_todos" "GET" "$BASE_URL/admin/dashboard/todos" "401" "false" "no token"
fi

# 8) Products list (optional public or auth depending)
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/products?limit=5" || true)
record "products_list" "GET" "$BASE_URL/products?limit=5" "$status" "$([[ "$status" == "200" || "$status" == "401" ]] && echo true || echo false)" "acceptable: 200 or 401 depending on auth"

# 9) Dashboard summary (need auth)
if [[ -n "$AUTH_H" ]]; then
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/admin/dashboard/summary" || true)
  record "admin_dashboard_summary" "GET" "$BASE_URL/admin/dashboard/summary" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""
else
  record "admin_dashboard_summary" "GET" "$BASE_URL/admin/dashboard/summary" "401" "false" "no token"
fi

# 10) Dashboard order trends (need auth)
if [[ -n "$AUTH_H" ]]; then
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/admin/dashboard/order-trends?days=7" || true)
  record "admin_dashboard_order_trends" "GET" "$BASE_URL/admin/dashboard/order-trends?days=7" "$status" "$( [[ "$status" == "200" ]] && echo true || echo false )" ""
else
  record "admin_dashboard_order_trends" "GET" "$BASE_URL/admin/dashboard/order-trends?days=7" "401" "false" "no token"
fi

# 11) Stores list (public or auth depending)
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/stores?limit=5" || true)
record "stores_list" "GET" "$BASE_URL/stores?limit=5" "$status" "$([[ "$status" == "200" || "$status" == "401" ]] && echo true || echo false)" "acceptable: 200 or 401 depending on auth"

# 12) Try get first order id and fetch detail (need auth)
order_id=""
if [[ -n "$AUTH_H" ]]; then
  orders_json=$(curl -s -H "$AUTH_H" "$BASE_URL/admin/orders?limit=1" || true)
  # 支持多种结构：{data:[...]}, {data:{list:[...]}}, {list:[...]}
  order_id=$(echo "$orders_json" | jq -r '.data[0].id // .data.list[0].id // .list[0].id // empty' 2>/dev/null || echo "")
  if [[ -n "$order_id" && "$order_id" != "null" ]]; then
    status=$(curl -s -o /dev/null -w "%{http_code}" -H "$AUTH_H" "$BASE_URL/admin/orders/$order_id" || true)
    record "admin_order_detail" "GET" "$BASE_URL/admin/orders/$order_id" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" "id=$order_id"
  else
    record "admin_order_detail" "GET" "$BASE_URL/admin/orders/{id}" "404" "false" "no orders found"
  fi
else
  record "admin_order_detail" "GET" "$BASE_URL/admin/orders/{id}" "401" "false" "no token"
fi

# Summary
total=$((pass_count + fail_count))
pass_rate=0
if [[ "$total" -gt 0 ]]; then
  pass_rate=$(python3 - <<PY
p=$pass_count
t=$total
print(round(p*100.0/t,2))
PY
)
fi
{
  echo "Total: $total"
  echo "Pass: $pass_count"
  echo "Fail: $fail_count"
  echo "PassRate: ${pass_rate}%"
} > "$SUMMARY_FILE"

echo "Summary written to $SUMMARY_FILE"
echo "Detail written to $REPORT_FILE"
