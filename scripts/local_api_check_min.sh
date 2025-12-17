#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BASE_URL=${BASE_URL:-"http://127.0.0.1:9292"}
REPORT_DIR=${REPORT_DIR:-"$ROOT_DIR/build-ci-logs"}
REPORT_FILE="$REPORT_DIR/local_api_check.csv"
mkdir -p "$REPORT_DIR"
if [[ ! -f "$REPORT_FILE" ]]; then echo "case,method,url,status,ok,notes" > "$REPORT_FILE"; fi
record(){ echo "$1,$2,$3,$4,$5,$6" >> "$REPORT_FILE"; }

# Dev login
resp=$(curl -sS -X POST -H 'Content-Type: application/json' -d '{"open_id":"user_openid_local_stateful","phone":"18880000001","nickname":"StatefulUser"}' "$BASE_URL/api/v1/user/dev-login" || true)
TOKEN=$(python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 t=(d.get('data') or {}).get('token') or d.get('token') or ''
 print(t)
except: print('')
PY
<<< "$resp")
AUTH_H=""; [[ -n "$TOKEN" ]] && AUTH_H="Authorization: Bearer $TOKEN"

# product_detail
pstatus=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/api/v1/products?page=1&size=1" || true)
prod_id=$(curl -s "$BASE_URL/api/v1/products?page=1&size=1" | python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 items=d.get('data') or []
 print(items[0].get('id') if items else '')
except: print('')
PY
)
if [[ -n "$prod_id" ]]; then
  dstatus=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/api/v1/products/$prod_id" || true)
  pok=$(curl -s "$BASE_URL/api/v1/products/$prod_id" | python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 print('true' if isinstance(d.get('sku'), list) else 'false')
except: print('false')
PY
)
  ok_combined=false; [[ "$dstatus" == "200" && "$pok" == "true" ]] && ok_combined=true
  record "product_detail" "GET" "$BASE_URL/products/$prod_id" "$dstatus" "$ok_combined" "sku_list=$pok"
else
  record "product_detail" "GET" "$BASE_URL/products/{id}" "$pstatus" "false" "no products"
fi

# cart_items
# fetch a sku_id from product detail if not yet
if [[ -z "${pok:-}" || "${pok}" != "true" ]]; then
  pd_json=$(curl -s "$BASE_URL/api/v1/products/$prod_id" || true)
  sku_id=$(python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 skus=d.get('sku') or d.get('skus') or []
 print((skus[0] or {}).get('id') if skus else '')
except: print('')
PY
<<< "$pd_json")
fi
add_body=$(python3 - <<'PY'
import os,json
pid=os.environ.get('PID','0')
sid=os.environ.get('SID','')
sk=os.environ.get('SK','')
payload={"product_id":int(pid or '0'),"quantity":1}
if sid:
 payload["store_id"]=int(sid)
if sk:
 payload["sku_id"]=int(sk)
print(json.dumps(payload))
PY
)
PID="$prod_id" SID="$store_id" SK="$sku_id" \
post_status=$(curl -s -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' ${AUTH_H:+-H "$AUTH_H"} -d "$add_body" "$BASE_URL/api/v1/cart" || true)
get_status=$(curl -s -o /dev/null -w '%{http_code}' ${AUTH_H:+-H "$AUTH_H"} "$BASE_URL/api/v1/cart" || true)
items_list=$(curl -s ${AUTH_H:+-H "$AUTH_H"} "$BASE_URL/api/v1/cart" | python3 - <<'PY'
import sys,json
try:
 c=json.loads(sys.stdin.read());
 print('true' if isinstance(c.get('items'), list) else 'false')
except: print('false')
PY
)
ok_combined=false; [[ "$get_status" == "200" && "$items_list" == "true" ]] && ok_combined=true
record "cart_items" "GET" "$BASE_URL/cart" "$get_status" "$ok_combined" "items_list=$items_list;post=$post_status"

# order_create
order_body=$(curl -s ${AUTH_H:+-H "$AUTH_H"} "$BASE_URL/api/v1/cart" | python3 - <<'PY'
import sys,json
try:
 c=json.loads(sys.stdin.read());
 items=[]
 for it in (c.get('items') or []):
  items.append({"product_id":it.get('product_id'),"sku_id":it.get('sku_id'),"qty":it.get('qty') or 1})
 print(json.dumps({"user_id":1,"items":items,"delivery_type":"store","pay_method":"wechat"}))
except: print('{"user_id":1,"items":[],"delivery_type":"store","pay_method":"wechat"}')
PY
)
oc_status=$(curl -s -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' ${AUTH_H:+-H "$AUTH_H"} -d "$order_body" "$BASE_URL/api/v1/orders" || true)
oid=$(curl -s -H 'Content-Type: application/json' ${AUTH_H:+-H "$AUTH_H"} -d "$order_body" "$BASE_URL/api/v1/orders" | python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 oid=d.get('order_id') or ((d.get('data') or {}).get('order_id'))
 print(oid or '')
except: print('')
PY
)
ok_combined=false; [[ ("$oc_status" == "200" || "$oc_status" == "201") && -n "$oid" ]] && ok_combined=true
record "order_create" "POST" "$BASE_URL/orders" "$oc_status" "$ok_combined" "order_id=${oid:-}"

# store_detail
sstatus=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/api/v1/stores?page=1&size=1" || true)
store_id=$(curl -s "$BASE_URL/api/v1/stores?page=1&size=1" | python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 items=d.get('data') or []
 print(items[0].get('id') if items else '')
except: print('')
PY
)
if [[ -n "$store_id" ]]; then
  sd_status=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/api/v1/stores/$store_id" || true)
  menus_list=$(curl -s "$BASE_URL/api/v1/stores/$store_id" | python3 - <<'PY'
import sys,json
try:
 d=json.loads(sys.stdin.read());
 print('true' if isinstance(d.get('menus'), list) else 'false')
except: print('false')
PY
)
  ok_combined=false; [[ "$sd_status" == "200" && "$menus_list" == "true" ]] && ok_combined=true
  record "store_detail" "GET" "$BASE_URL/stores/$store_id" "$sd_status" "$ok_combined" "menus_list=$menus_list"
else
  record "store_detail" "GET" "$BASE_URL/stores/{id}" "$sstatus" "false" "no stores"
fi
echo "Done. CSV at $REPORT_FILE"
