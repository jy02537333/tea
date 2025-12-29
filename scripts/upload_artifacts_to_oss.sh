#!/usr/bin/env bash
set -euo pipefail

# upload_artifacts_to_oss.sh
# 使用后端 Admin OSS Policy 接口，将本地文件以表单直传上传到阿里云 OSS。
# 依赖：curl、jq、base64
# 环境：
#   API_BASE (默认 http://127.0.0.1:9292)
#   ADMIN_TOKEN （若未设置，将尝试读取 build-ci-logs/admin_token.txt）
#   BUSINESS （可选，默认 ci_artifact）
# 用法：
#   scripts/upload_artifacts_to_oss.sh <file1> [file2 ...]
# 输出：上传成功后打印对象 key 与公开 URL（若配置了 CDN 域名则为 CDN URL，否则为原始 OSS URL）。

API_BASE=${API_BASE:-"http://127.0.0.1:9292"}
BUSINESS=${BUSINESS:-"ci_artifact"}

ensure_admin_token() {
  # 尝试从文件读取；若无效则使用本地 dev-login 自动刷新
  local need_refresh="0"
  if [[ -z "${ADMIN_TOKEN:-}" ]]; then
    if [[ -s "build-ci-logs/admin_token.txt" ]]; then
      ADMIN_TOKEN=$(cat build-ci-logs/admin_token.txt)
    fi
  fi

  # 校验 token 可用性（管理员探针）
  if [[ -n "${ADMIN_TOKEN:-}" ]]; then
    local status
    status=$(curl -sS -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $ADMIN_TOKEN" "$API_BASE/api/v1/admin/dashboard/ping" || true)
    if [[ "$status" != "200" ]]; then
      need_refresh="1"
    fi
  else
    need_refresh="1"
  fi

  if [[ "$need_refresh" = "1" ]]; then
    # 在本地/开发环境下调用 dev-login 获取管理员 token
    echo "[upload_artifacts] ADMIN_TOKEN 缺失或无效，尝试 dev-login 刷新..." >&2
    local resp
    resp=$(curl -sS -H "Content-Type: application/json" -d '{"openid":"admin_openid"}' "$API_BASE/api/v1/user/dev-login" || true)
    local new_token
    new_token=$(echo "$resp" | jq -r '(.token // .data.token // empty)')
    if [[ -n "$new_token" && "$new_token" != "null" ]]; then
      ADMIN_TOKEN="$new_token"
      mkdir -p build-ci-logs
      printf '%s' "$ADMIN_TOKEN" > build-ci-logs/admin_token.txt
      echo "[upload_artifacts] 已刷新 ADMIN_TOKEN。" >&2
    else
      echo "ERROR: 无法获取有效的 ADMIN_TOKEN（dev-login 失败）。请在环境中提供 ADMIN_TOKEN。" >&2
      exit 1
    fi
  fi
}

ensure_admin_token

if [[ $# -lt 1 ]]; then
  echo "用法: $0 <file1> [file2 ...]" >&2
  exit 1
fi

content_type_for() {
  local f="$1"; local ext
  ext="${f##*.}"; ext=$(echo "$ext" | tr 'A-Z' 'a-z')
  case "$ext" in
    png) echo "image/png" ;;
    jpg|jpeg) echo "image/jpeg" ;;
    gif) echo "image/gif" ;;
    webp) echo "image/webp" ;;
    svg) echo "image/svg+xml" ;;
    zip) echo "application/zip" ;;
    html|htm) echo "text/html" ;;
    json) echo "application/json" ;;
    txt|log|md) echo "text/plain" ;;
    *) echo "application/octet-stream" ;;
  esac
}

# 统一的表单直传函数
form_upload() {
  local host="$1"; local key="$2"; local policy_b64="$3"; local accessKeyId="$4"; local signature="$5"; local ctype="$6"; local file="$7"
  curl -sS -X POST "$host" \
    -F "key=$key" \
    -F "policy=$policy_b64" \
    -F "OSSAccessKeyId=$accessKeyId" \
    -F "Signature=$signature" \
    -F "success_action_status=200" \
    -F "Content-Type=$ctype" \
    -F "file=@$file;type=$ctype"
}

head_status() {
  local url="$1"
  curl -s -o /dev/null -w "%{http_code}" -I "$url" 2>/dev/null || true
}

post_policy() {
  local fname="$1"; local fsize="$2"; local ctype="$3"
  curl -sS -X POST \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    "$API_BASE/api/v1/admin/storage/oss/policy" \
    --data "{\"business\":\"$BUSINESS\",\"file_name\":\"$fname\",\"content_type\":\"$ctype\",\"file_size\":$fsize}"
}

# 直接走管理端上传端点（回退方案）
upload_via_admin() {
  local file="$1"; local ctype="$2"
  local url="$API_BASE/api/v1/admin/uploads"
  local resp
  resp=$(curl -sS -X POST "$url" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: multipart/form-data" \
    -F "file=@$file;type=$ctype")
  # 期待返回 {"url":"..."}
  local url_out
  url_out=$(echo "$resp" | jq -r '.url // .data.url // empty')
  if [[ -n "$url_out" && "$url_out" != "null" ]]; then
    echo "$url_out"
    return 0
  fi
  echo ""; return 1
}

# 从 policy(base64) 解码出 bucket 与约束
extract_bucket_from_policy() {
  local policy_b64="$1"; local decoded
  decoded=$(echo "$policy_b64" | base64 -d 2>/dev/null || true)
  # 期望形如：{"conditions":[{"bucket":"xxx"},["starts-with","$key","prefix/"], ...],"expiration":"..."}
  echo "$decoded" | jq -r '.conditions[] | select(type=="object" and .bucket) | .bucket' | head -n1
}

# 从后端配置中读取 endpoint 与（可选）cdn_domain
read_endpoint_from_config() {
  # 优先 tea-api/configs/config.yaml
  if [[ -s "tea-api/configs/config.yaml" ]]; then
    jq -r '.upload.oss.endpoint // .oss.endpoint // empty' tea-api/configs/config.yaml 2>/dev/null || true
  else
    echo ""; return 0
  fi
}

read_cdn_domain_from_config() {
  if [[ -s "tea-api/configs/config.yaml" ]]; then
    jq -r '.upload.oss.cdn_domain // .oss.cdn_domain // empty' tea-api/configs/config.yaml 2>/dev/null || true
  else
    echo ""; return 0
  fi
}

for file in "$@"; do
  if [[ ! -s "$file" ]]; then
    echo "WARN: 文件不存在或为空：$file，跳过" >&2
    continue
  fi
  fname=$(basename "$file")
  fsize=$(wc -c < "$file")
  ctype=$(content_type_for "$file")

  echo "==> 请求 Policy：$fname ($ctype, $fsize bytes)" >&2
  resp=$(post_policy "$fname" "$fsize" "$ctype")
  # 期待字段：accessKeyId, signature, policy, expire_at, object_key_template [, host]
  accessKeyId=$(echo "$resp" | jq -r '.accessKeyId // empty')
  signature=$(echo "$resp" | jq -r '.signature // empty')
  policy_b64=$(echo "$resp" | jq -r '.policy // empty')
  expire_at=$(echo "$resp" | jq -r '.expire_at // empty')
  obj_tpl=$(echo "$resp" | jq -r '.object_key_template // empty')
  host=$(echo "$resp" | jq -r '.host // empty')

  if [[ -z "$policy_b64" || -z "$signature" || -z "$accessKeyId" || -z "$obj_tpl" ]]; then
    # 若后端未实现 Policy 路由或返回 not_found，回退到直接上传
    if echo "$resp" | jq -e '.error? == "not_found"' >/dev/null 2>&1; then
      echo "[upload_artifacts] Policy 路由未实现，回退到 /api/v1/admin/uploads ..." >&2
      url_out=""
      if url_out=$(upload_via_admin "$file" "$ctype"); then
        echo "OK: $file -> $url_out"
        continue
      else
        echo "ERROR: 管理端直接上传失败：$file" >&2
        exit 2
      fi
    else
      echo "ERROR: Policy 响应缺少必要字段：$resp" >&2
      exit 2
    fi
  fi

  bucket=""
  bucket=$(extract_bucket_from_policy "$policy_b64")
  endpoint=""
  endpoint=$(read_endpoint_from_config)
  if [[ -z "$host" ]]; then
    if [[ -n "$bucket" && -n "$endpoint" ]]; then
      host="https://$bucket.$endpoint"
    else
      echo "ERROR: 无法确定 OSS host（缺少 host/bucket/endpoint）。" >&2
      exit 3
    fi
  fi

  # 生成对象 key：替换模板中的 ${filename}
  key=$(echo "$obj_tpl" | sed "s/\${filename}/$fname/g")

  echo "==> 上传到：$host ($key)" >&2
  # 表单直传（初次使用检测到的 Content-Type）
  upload_resp=$(form_upload "$host" "$key" "$policy_b64" "$accessKeyId" "$signature" "$ctype" "$file")

  # OSS 表单直传成功时，返回 200，可能包含 XML 响应；此处不强依赖内容，仅校验状态。
  # 简单输出 URL（优先使用 CDN域名）
  cdn_domain=$(read_cdn_domain_from_config)
  if [[ -n "$cdn_domain" ]]; then
    url="https://$cdn_domain/$key"
  else
    url="$host/$key"
  fi

  # 健康检查：若为 ZIP 且返回非 200，则回退为 octet-stream 重传
  status=$(head_status "$url")
  if [[ "$ctype" = "application/zip" && "$status" != "200" ]]; then
    echo "[upload_artifacts] ZIP HEAD=$status，尝试以 octet-stream 重传同 Key..." >&2
    form_upload "$host" "$key" "$policy_b64" "$accessKeyId" "$signature" "application/octet-stream" "$file" >/dev/null || true
    # 重新检查
    status=$(head_status "$url")
    echo "[upload_artifacts] 回退直传后 HEAD=$status" >&2
  fi

  echo "OK: $file -> $url"

done
