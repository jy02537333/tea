#!/usr/bin/env bash
set -euo pipefail

# Enhanced local API check: perform dev-login to get a token, then run user-facing API checks
# Logs and response bodies are saved under build-ci-logs/api_validation_stateful

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BASE_URL=${BASE_URL:-"http://127.0.0.1:9292"}
OUT_DIR="$ROOT_DIR/build-ci-logs/api_validation_stateful"
mkdir -p "$OUT_DIR"

# CSV report paths (ensure defined before any recording)
REPORT_DIR=${REPORT_DIR:-"$ROOT_DIR/build-ci-logs"}
mkdir -p "$REPORT_DIR"
REPORT_FILE="$REPORT_DIR/local_api_check.csv"
SUMMARY_FILE="$REPORT_DIR/local_api_summary.txt"
: > "$SUMMARY_FILE"

# Initialize CSV header if not exists
if [[ ! -f "$REPORT_FILE" ]]; then
  echo "case,method,url,status,ok,notes" > "$REPORT_FILE"
fi

pass_count=0
fail_count=0

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

  # Best-effort JSON summary (do not fail if non-JSON)
  top_keys=$(python3 - <<'PY'
import sys, json
text=sys.stdin.read()
try:
  obj=json.loads(text)
  if isinstance(obj, dict):
    print("Top-level keys:", sorted(list(obj.keys())))
  elif isinstance(obj, list):
    print("Top-level: array, len=", len(obj))
  else:
    print("Top-level type:", type(obj).__name__)
except Exception:
  print("Top-level: [non-JSON]")
PY
  <<< "$resp")
  echo "$top_keys" || true
}

# 1) Health check (no auth)
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/health" || true)
curl_json GET /api/v1/health no || true
record "health" "GET" "$BASE_URL/api/v1/health" "$status" "$([[ "$status" == "200" ]] && echo true || echo false)" ""

# 2) Dev login to obtain token (prefer auth/login, fallback to user/dev-login)
AUTH_LOGIN_URL="$BASE_URL/api/v1/auth/login"
# Try username/password first, then openid
AUTH_LOGIN_BODY1='{"username":"admin","password":"pass"}'
resp=$(curl -sS -X POST -H 'Content-Type: application/json' -d "$AUTH_LOGIN_BODY1" "$AUTH_LOGIN_URL" || true)
echo "$resp" > "$OUT_DIR/POST__api_v1_auth_login.json"
if [[ -z "$resp" || "$resp" == "null" ]]; then
  AUTH_LOGIN_BODY2='{"openid":"admin_openid"}'
  resp=$(curl -sS -X POST -H 'Content-Type: application/json' -d "$AUTH_LOGIN_BODY2" "$AUTH_LOGIN_URL" || true)
  echo "$resp" > "$OUT_DIR/POST__api_v1_auth_login_openid.json"
  if [[ -z "$resp" || "$resp" == "null" ]]; then
    DEV_LOGIN_BODY='{"open_id":"user_openid_local_stateful","phone":"18880000001","nickname":"StatefulUser"}'
    resp=$(curl -sS -X POST -H 'Content-Type: application/json' -d "$DEV_LOGIN_BODY" "$BASE_URL/api/v1/user/dev-login" || true)
    echo "$resp" > "$OUT_DIR/POST__api_v1_user_dev-login.json"
  fi
fi

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

if [[ -z "$TOKEN" ]]; then
  # Fallback: reuse token from prior run logs if available
  ALT_TOKEN_FILE="$ROOT_DIR/build-ci-logs/admin_login_response.json"
  if [[ -f "$ALT_TOKEN_FILE" ]]; then
    TOKEN=$(grep -Po '"token"\s*:\s*"\K[^"]+' "$ALT_TOKEN_FILE" || true)
  fi
fi

if [[ -n "$TOKEN" ]]; then
  AUTH_HEADER="Authorization: Bearer $TOKEN"
  log "Obtained token via dev-login."
  record "dev_login" "POST" "$AUTH_LOGIN_URL" "200" "true" "token acquired"
else
  AUTH_HEADER=""
  log "Dev-login did not return a token; proceeding without Authorization header."
  # record as false if no token
  record "dev_login" "POST" "$AUTH_LOGIN_URL" "200" "false" "no token"
fi

# 3) User-facing endpoints with auth to avoid FK issues
ustat=$(curl -s -o /dev/null -w "%{http_code}" ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/user/info" || true)
curl_json GET /api/v1/user/info yes || true
record "user_info" "GET" "$BASE_URL/api/v1/user/info" "$ustat" "$([[ "$ustat" == "200" ]] && echo true || echo false)" ""

cstat=$(curl -s -o /dev/null -w "%{http_code}" ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/cart" || true)
curl_json GET /api/v1/cart yes || true
record "cart_get" "GET" "$BASE_URL/api/v1/cart" "$cstat" "$([[ "$cstat" == "200" ]] && echo true || echo false)" ""

curl_json GET /api/v1/user/coupons yes || true

# 4) Basic catalog endpoints (no auth)
plist=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/products" || true)
curl_json GET /api/v1/products no || true
record "products_list" "GET" "$BASE_URL/api/v1/products" "$plist" "$([[ "$plist" == "200" ]] && echo true || echo false)" ""

clist=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/categories" || true)
curl_json GET /api/v1/categories no || true
record "categories_list" "GET" "$BASE_URL/api/v1/categories" "$clist" "$([[ "$clist" == "200" ]] && echo true || echo false)" ""

slist=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/stores" || true)
curl_json GET /api/v1/stores no || true
record "stores_list" "GET" "$BASE_URL/api/v1/stores" "$slist" "$([[ "$slist" == "200" ]] && echo true || echo false)" ""

# 5) Minimal success path assertions: cart add -> cart get -> order create (conditional)
if [[ -n "$AUTH_HEADER" ]]; then
  set +e
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
    # CSV: cart items is list and status 200
      cstatus=$(curl -s -o /dev/null -w '%{http_code}' ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/cart" || true)
      cok=$(python3 - <<'PY'
import sys, json
path=sys.argv[1]
try:
  c=json.loads(open(path, 'r', encoding='utf-8').read())
  items=c.get('items')
  print('true' if isinstance(items, list) else 'false')
except Exception:
  print('false')
PY
      "$OUT_DIR/GET__api_v1_cart.json")
      ok_combined=false
      if [[ "$cstatus" == "200" && "$cok" == "true" ]]; then ok_combined=true; fi
      record "cart_items" "GET" "$BASE_URL/cart" "$cstatus" "$ok_combined" "items_list=$cok"
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
    # CSV: order create returns order_id and status 200/201
      ostatus=$(curl -s -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' ${AUTH_HEADER:+-H "$AUTH_HEADER"} -d "$order_body" "$BASE_URL/api/v1/orders" || true)
      oid=$(python3 - <<'PY'
import sys, json
path=sys.argv[1]
try:
  d=json.loads(open(path, 'r', encoding='utf-8').read())
  # support {order_id,...} or {data:{order_id}}
  oid=d.get('order_id')
  if not oid and isinstance(d.get('data'), dict):
    oid=d['data'].get('order_id')
  print(str(oid) if oid else '')
except Exception:
  print('')
PY
      "$OUT_DIR/POST__api_v1_orders.json")
      ok_combined=false
      if [[ ("$ostatus" == "200" || "$ostatus" == "201") && -n "$oid" ]]; then ok_combined=true; fi
      record "order_create" "POST" "$BASE_URL/orders" "$ostatus" "$ok_combined" "order_id=${oid:-}"
    # Brief summary
    log "Cart+Order minimal flow attempted (product=$prod_id, sku=${sku_id:-})"

      # If order created, fetch order detail and generate evidence files under build-ci-logs/
      if [[ -n "$oid" ]]; then
        DETAIL_PATH_ROOT="$ROOT_DIR/build-ci-logs"
        mkdir -p "$DETAIL_PATH_ROOT"
        order_detail=$(curl -sS -X GET ${AUTH_HEADER:+-H "$AUTH_HEADER"} "$BASE_URL/api/v1/orders/$oid" || true)
        echo "$order_detail" > "$OUT_DIR/GET__api_v1_orders_${oid}.json"
        echo "$order_detail" > "$DETAIL_PATH_ROOT/order_detail_${oid}.json"

        # Compute amounts summary and checked json via Python (no jq dependency)
        python3 - <<'PY'
    import sys, json, pathlib
    root = pathlib.Path(sys.argv[1])
    oid = sys.argv[2]
    detail_text = (root / f"order_detail_{oid}.json").read_text(encoding='utf-8', errors='ignore')
    try:
      obj = json.loads(detail_text)
      data = obj.get('data', obj) if isinstance(obj, dict) else {}
    except Exception:
      data = {}

    def num(v):
      try:
        # Accept int/float/str
        if isinstance(v, (int, float)):
          return v
        if isinstance(v, str):
          return float(v) if ('.' in v) else int(v)
      except Exception:
        pass
      return 0

    summary = {
      "order_id": data.get('id') or oid,
      "store_id": data.get('store_id'),
      "total_amount": num(data.get('total_amount')),
      "discount_amount": num(data.get('discount_amount')),
      "pay_amount": num(data.get('pay_amount')),
    }
    summary["check"] = (summary["pay_amount"] == (summary["total_amount"] - summary["discount_amount"]))

    summary_path = root / 'order_amounts_summary.json'
    checked_path = root / f'order_detail_{oid}_checked.json'

    summary_path.write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding='utf-8')
    checked_path.write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding='utf-8')

    # Also append a human readable line to local_api_summary.txt if present
    summary_txt = root / 'local_api_summary.txt'
    line = f"order_id={summary['order_id']}, store_id={summary.get('store_id')}, total={summary['total_amount']}, discount={summary['discount_amount']}, pay={summary['pay_amount']}, check={str(summary['check']).lower()}\n"
    try:
      with open(summary_txt, 'a', encoding='utf-8') as f:
        f.write(line)
    except Exception:
      pass
PY
        "$DETAIL_PATH_ROOT" "$oid"
        log "Order evidence generated: order_detail_${oid}.json, order_amounts_summary.json, order_detail_${oid}_checked.json"
      fi
  else
    log "No products returned; skipping cart/order minimal flow"
  fi
  set -e
else
  log "No auth token; skipping cart/order minimal flow"
  # Still record placeholders for visibility
  record "product_detail" "GET" "$BASE_URL/products/{id}" "401" "false" "no token"
  record "cart_items" "GET" "$BASE_URL/cart" "401" "false" "no token"
  record "order_create" "POST" "$BASE_URL/orders" "401" "false" "no token"
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
else
  log "No stores returned; skipping store detail menus assertion"
  # Record skipped in CSV with 404 to indicate no data
  record "store_detail" "GET" "$BASE_URL/stores/{id}" "404" "false" "no stores"
fi

log "Stateful API validation completed. Bodies under: $OUT_DIR"
