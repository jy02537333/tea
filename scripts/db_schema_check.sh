#!/usr/bin/env bash
set -euo pipefail

# Simple DB schema existence and fields check for core tables.
# Env overrides
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3308}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-gs963852}"
DB_NAME="${DB_NAME:-tea_shop}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/build-ci-logs/api_validation"
mkdir -p "$OUT_DIR"
CHECKLIST="$OUT_DIR/db_checklist.txt"
SUMMARY_OUT="$OUT_DIR/db_check_summary.txt"

: > "$CHECKLIST"
: > "$SUMMARY_OUT"

mysql_cmd() {
  mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -N -e "$1" || return 1
}

echo "Database: $DB_HOST:$DB_PORT / $DB_NAME" >> "$SUMMARY_OUT"
echo >> "$SUMMARY_OUT"

tables=(
  users user_profiles roles user_roles stores store_staff categories products product_skus product_images store_products inventories cart_items addresses orders order_items payments coupons_templates coupons wallets wallet_transactions points_transactions membership_packages user_memberships referrals commissions commission_transactions platform_financial_settings print_jobs notifications activities activity_registrations shipping_templates shipping_rules admin_audit_logs permissions role_permissions partner_levels user_bank_accounts store_bank_accounts withdrawal_requests referrals_closure
)

for t in "${tables[@]}"; do
  if mysql_cmd "SHOW TABLES LIKE '$t';" | grep -q "^$t$"; then
    echo "[OK] table exists: $t" >> "$SUMMARY_OUT"
  else
    echo "[MISS] table missing: $t" >> "$SUMMARY_OUT"
  fi
done

echo >> "$SUMMARY_OUT"
echo "Field checks (key tables):" >> "$SUMMARY_OUT"

check_fields() {
  local table="$1"; shift
  local want=("$@")
  local cols
  cols=$(mysql_cmd "SHOW COLUMNS FROM \`$table\`;") || { echo "[ERR] cannot read columns for $table" >> "$SUMMARY_OUT"; return; }
  for w in "${want[@]}"; do
    if echo "$cols" | awk '{print $1}' | grep -q "^$w$"; then
      echo "  - $table.$w: OK" >> "$SUMMARY_OUT"
    else
      echo "  - $table.$w: MISSING" >> "$SUMMARY_OUT"
    fi
  done
}

check_fields users id phone status created_at updated_at
check_fields wallets user_id balance frozen
check_fields points_transactions id user_id change created_at
check_fields coupons id user_id status expires_at
check_fields membership_packages id name price type
check_fields user_memberships id user_id package_id status started_at expires_at

echo >> "$SUMMARY_OUT"
echo "Index/constraints quick checks:" >> "$SUMMARY_OUT"
quick_sqls=(
  "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema='$DB_NAME' AND table_name='users' AND index_name='uk_users_phone';"
  "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema='$DB_NAME' AND table_name='orders' AND index_name='uk_orders_order_no';"
)
labels=("users.uk_users_phone" "orders.uk_orders_order_no")

for i in $(seq 0 $((${#quick_sqls[@]}-1))); do
  cnt=$(mysql_cmd "${quick_sqls[$i]}" || echo 0)
  if [[ "${cnt:-0}" != "0" ]]; then
    echo "  - ${labels[$i]}: OK" >> "$SUMMARY_OUT"
  else
    echo "  - ${labels[$i]}: MISSING" >> "$SUMMARY_OUT"
  fi
done

cat > "$CHECKLIST" <<'TXT'
核心表与关键字段核对清单（依据 db/schema.sql）

- users: id, phone (unique), email, password_hash, status, created_at, updated_at
- wallets: user_id (PK, FK users.id), balance, frozen
- points_transactions: id, user_id, change, created_at
- coupons: id, user_id, status, expires_at, claimed_at, template_id (FK)
- membership_packages: id, name, price, type, discount_rate, purchase_discount_rate
- user_memberships: id, user_id (FK), package_id (FK), status, started_at, expires_at
- orders: id, order_no (unique), user_id (FK), type, status, total_amount, pay_amount, created_at
- order_items: id, order_id (FK), product_id, sku_id, quantity, unit_price, total_price
- payments: id, order_id (FK), channel, amount, status, payload
- permissions/role_permissions: code(unique), role→permission mapping

索引/约束：
- users.uk_users_phone
- orders.uk_orders_order_no

运维建议：以 db/schema.sql 为权威来源执行建表/迁移；任何结构变更需同步更新 doc/db_schema.md。
TXT

echo "DB schema check completed. See:" >> "$SUMMARY_OUT"
echo " - $SUMMARY_OUT" >> "$SUMMARY_OUT"
echo " - $CHECKLIST" >> "$SUMMARY_OUT"

echo "Schema check summary written to $SUMMARY_OUT"
echo "Checklist written to $CHECKLIST"
