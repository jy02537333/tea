#!/usr/bin/env bash
set -euo pipefail

# migration_for_dba.sh
# 说明: 给 DBA 的可复制脚本 — 在能访问 MySQL (127.0.0.1:3308) 的主机上运行
# 使用方法示例:
#   bash scripts/migration_for_dba.sh \
#     --host 127.0.0.1 --port 3308 --user root --password 'gs963852' --db tea_shop
# 或更安全的环境变量方式:
#   export DB_PW='gs963852'
#   bash scripts/migration_for_dba.sh --host 127.0.0.1 --port 3308 --user root --db tea_shop

show_usage() {
  cat <<EOF
Usage: $0 --host HOST --port PORT --user USER --db DB [--password PASSWORD] [--backup-dir DIR]

Options:
  --host       MySQL host (default: 127.0.0.1)
  --port       MySQL port (default: 3308)
  --user       MySQL user
  --password   MySQL password (optional; if omitted will read from DB_PW env or prompt)
  --db         Database name (e.g. tea_shop)
  --backup-dir Directory to write backups (default: build-ci-logs)
  --help       Show this help

Security note: passing password on the command line is insecure (visible in process list).
Prefer setting DB_PW environment variable or using an interactive prompt.
EOF
}

# defaults
HOST="127.0.0.1"
PORT="3308"
USER="root"
DB="tea_shop"
BACKUP_DIR="build-ci-logs"
PW=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) HOST="$2"; shift 2;;
    --port) PORT="$2"; shift 2;;
    --user) USER="$2"; shift 2;;
    --password) PW="$2"; shift 2;;
    --db) DB="$2"; shift 2;;
    --backup-dir) BACKUP_DIR="$2"; shift 2;;
    --help) show_usage; exit 0;;
    *) echo "Unknown arg: $1"; show_usage; exit 2;;
  esac
done

# prefer env var if password not provided
if [[ -z "$PW" ]]; then
  if [[ -n "${DB_PW:-}" ]]; then
    PW="$DB_PW"
  fi
fi

if [[ -z "$PW" ]]; then
  # prompt (no-echo)
  read -s -p "MySQL password for ${USER}@${HOST}:${PORT}: " PW
  echo
fi

mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/tea_shop_backup_$(date +%F_%H%M%S).sql"
MIGRATION_SQL="tea-api/sql迁移.sql"
LOG_PREFIX="$BACKUP_DIR/migration"

# 1) Backup
echo "[1/4] Creating backup to $BACKUP_FILE"
mysqldump -h "$HOST" -P "$PORT" -u "$USER" -p"$PW" --databases "$DB" > "$BACKUP_FILE"
if [[ ! -s "$BACKUP_FILE" ]]; then
  echo "Backup file is empty or missing: $BACKUP_FILE" >&2
  exit 3
fi

echo "Backup created: $BACKUP_FILE"

# 2) (Optional) Check backup integrity (basic)
echo "[2/4] Checking backup integrity (simple gzip test)"
if command -v gzip >/dev/null 2>&1; then
  gzip -t <(cat "$BACKUP_FILE") >/dev/null 2>&1 || echo "Warning: gzip test not available for plain SQL"
else
  echo "gzip not available; skipping gzip integrity test"
fi

# 3) Apply migration
if [[ ! -f "$MIGRATION_SQL" ]]; then
  echo "Migration file not found: $MIGRATION_SQL" >&2
  exit 4
fi

echo "[3/4] Applying migration: $MIGRATION_SQL"
# save stdout/stderr
MYSQL_CMD=(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PW" "$DB")
"
"${MYSQL_CMD[@]}" < "$MIGRATION_SQL" 2>&1 | tee "$LOG_PREFIX"_apply_$(date +%F_%H%M%S).log

# 4) Verify
echo "[4/4] Verifying tables post-migration"
"${MYSQL_CMD[@]}" -e "USE $DB; SHOW TABLES\G" | tee "$BACKUP_DIR/migration_verify_after.txt"

echo "Migration script finished. Logs and verification written to $BACKUP_DIR"

echo "If you need to rollback, you can restore the backup with:"
echo "  mysql -h $HOST -P $PORT -u $USER -p'****' < $BACKUP_FILE"

echo "Note: replace '****' with the password or use DB_PW env var to avoid putting password on command line."
