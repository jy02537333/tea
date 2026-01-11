#!/usr/bin/env bash
set -eu

# One-click launcher for tea-api with configurable Redis and RabbitMQ.
# It wraps the existing ./run-tea-api.sh and passes environment overrides.
#
# Usage:
#   ./start_tea_api.sh \
#     --redis-host 127.0.0.1 --redis-port 6379 --redis-password 123456 \
#     --rabbitmq-host 127.0.0.1 --rabbitmq-port 5672 \
#     --rabbitmq-username guest --rabbitmq-password guest --rabbitmq-vhost / \
#     [--server-port 9292] [--jwt-secret dev_secret_change_me]
#   # or use URLs:
#   ./start_tea_api.sh --redis-url redis://:123456@192.168.3.82:6379/0 \
#                      --rabbitmq-url amqp://guest:guest@192.168.3.82:5672/
#
# All flags are optional; sensible defaults are provided.

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

# Defaults
REDIS_HOST="127.0.0.1"
REDIS_PORT="6379"
REDIS_PASSWORD="123456"

RABBITMQ_HOST="127.0.0.1"
RABBITMQ_PORT="5672"
RABBITMQ_USERNAME="guest"
RABBITMQ_PASSWORD="guest"
RABBITMQ_VHOST="/"

SERVER_PORT="9292"
JWT_SECRET="tea-shop-jwt-secret-key-2023"
DATABASE_DSN=""

# Helpers to parse URLs
parse_redis_url() {
  # redis://[:password@]host[:port][/db]
  local url="$1"
  url="${url#redis://}"
  local creds="" hostportpath="$url"
  if [[ "$url" == *@* ]]; then
    creds="${url%%@*}"
    hostportpath="${url#*@}"
  fi
  # password
  if [[ -n "$creds" ]]; then
    REDIS_PASSWORD="${creds#:}"
  fi
  local hostport="${hostportpath%%/*}"
  local dbpath="${hostportpath#*/}"
  if [[ "$dbpath" != "$hostportpath" ]]; then
    # has /db
    REDIS_DB="${dbpath}" || true
  fi
  local host="${hostport%%:*}"
  local port="${hostport#*:}"
  if [[ "$port" == "$host" ]]; then port="6379"; fi
  REDIS_HOST="$host"
  REDIS_PORT="$port"
}

parse_rabbitmq_url() {
  # amqp://user:pass@host:port/vhost
  local url="$1"
  url="${url#amqp://}"
  local userpass="${url%%@*}"
  local hostportpath="${url#*@}"
  local user="${userpass%%:*}"
  local pass="${userpass#*:}"
  local hostport="${hostportpath%%/*}"
  local vpath="${hostportpath#*/}"
  local host="${hostport%%:*}"
  local port="${hostport#*:}"
  if [[ "$port" == "$host" ]]; then port="5672"; fi
  RABBITMQ_USERNAME="$user"
  RABBITMQ_PASSWORD="$pass"
  RABBITMQ_HOST="$host"
  RABBITMQ_PORT="$port"
  if [[ -z "$vpath" ]]; then
    RABBITMQ_VHOST="/"
  else
    [[ "$vpath" == /* ]] && RABBITMQ_VHOST="$vpath" || RABBITMQ_VHOST="/$vpath"
  fi
}

# Parse flags (support --flag value and --flag=value forms)
while [ $# -gt 0 ]; do
  case "$1" in
    --redis-url=*) parse_redis_url "${1#*=}"; shift ;;
    --redis-url) parse_redis_url "${2:?missing value for --redis-url}"; shift 2 ;;

    --redis-host=*) REDIS_HOST="${1#*=}"; shift ;;
    --redis-host) REDIS_HOST="${2:?missing value for --redis-host}"; shift 2 ;;

    --redis-port=*) REDIS_PORT="${1#*=}"; shift ;;
    --redis-port) REDIS_PORT="${2:?missing value for --redis-port}"; shift 2 ;;

    --redis-password=*) REDIS_PASSWORD="${1#*=}"; shift ;;
    --redis-password) REDIS_PASSWORD="${2:?missing value for --redis-password}"; shift 2 ;;

    --rabbitmq-url=*) parse_rabbitmq_url "${1#*=}"; shift ;;
    --rabbitmq-url) parse_rabbitmq_url "${2:?missing value for --rabbitmq-url}"; shift 2 ;;

    --rabbitmq-host=*) RABBITMQ_HOST="${1#*=}"; shift ;;
    --rabbitmq-host) RABBITMQ_HOST="${2:?missing value for --rabbitmq-host}"; shift 2 ;;

    --rabbitmq-port=*) RABBITMQ_PORT="${1#*=}"; shift ;;
    --rabbitmq-port) RABBITMQ_PORT="${2:?missing value for --rabbitmq-port}"; shift 2 ;;

    --rabbitmq-username=*) RABBITMQ_USERNAME="${1#*=}"; shift ;;
    --rabbitmq-username) RABBITMQ_USERNAME="${2:?missing value for --rabbitmq-username}"; shift 2 ;;

    --rabbitmq-password=*) RABBITMQ_PASSWORD="${1#*=}"; shift ;;
    --rabbitmq-password) RABBITMQ_PASSWORD="${2:?missing value for --rabbitmq-password}"; shift 2 ;;

    # rabbitmq vhost: allow omitting value to keep default
    --rabbitmq-vhost=*) RABBITMQ_VHOST="${1#*=}"; shift ;;
    --rabbitmq-vhost)
      case "${2-}" in
        --*|"") shift 1 ;; # no value provided, keep default
        *) RABBITMQ_VHOST="${2}"; shift 2 ;;
      esac ;;

    --server-port=*) SERVER_PORT="${1#*=}"; shift ;;
    --server-port) SERVER_PORT="${2:?missing value for --server-port}"; shift 2 ;;

    --jwt-secret=*) JWT_SECRET="${1#*=}"; shift ;;
    --jwt-secret) JWT_SECRET="${2:?missing value for --jwt-secret}"; shift 2 ;;

    --dsn=*) DATABASE_DSN="${1#*=}"; shift ;;
    --dsn) DATABASE_DSN="${2:?missing value for --dsn}"; shift 2 ;;

    -h|--help)
      sed -n '1,40p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "Unknown flag: $1" >&2; exit 1 ;;
  esac
done

# Export environment overrides recognized by tea-api
export TEA_REDIS_HOST="$REDIS_HOST"
export TEA_REDIS_PORT="$REDIS_PORT"
export TEA_REDIS_PASSWORD="$REDIS_PASSWORD"

export TEA_RABBITMQ_HOST="$RABBITMQ_HOST"
export TEA_RABBITMQ_PORT="$RABBITMQ_PORT"
export TEA_RABBITMQ_USERNAME="$RABBITMQ_USERNAME"
export TEA_RABBITMQ_PASSWORD="$RABBITMQ_PASSWORD"
export TEA_RABBITMQ_VHOST="$RABBITMQ_VHOST"

# Optional server port + JWT
export TEA_SERVER_PORT="$SERVER_PORT"
export TEA_JWT_SECRET="$JWT_SECRET"
if [[ -n "$DATABASE_DSN" ]]; then
  export TEA_DSN="$DATABASE_DSN"
fi

# Summary
echo "== tea-api start parameters =="
echo "Redis:     $REDIS_HOST:$REDIS_PORT (pwd set: $([[ -n "$REDIS_PASSWORD" ]] && echo yes || echo no))"
echo "RabbitMQ:  $RABBITMQ_USERNAME@$RABBITMQ_HOST:$RABBITMQ_PORT vhost=$RABBITMQ_VHOST"
echo "Server:    port=$SERVER_PORT"
echo "JWT:       len=${#JWT_SECRET}"
echo "Database:  ${DATABASE_DSN:+(DSN provided)}${DATABASE_DSN:+ }"

# Start via existing unified launcher (handles build, port checks, health, token)
chmod +x ./run-tea-api.sh
./run-tea-api.sh
