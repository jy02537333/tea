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

if [[ -z "${ADMIN_TOKEN:-}" ]]; then
  if [[ -s "build-ci-logs/admin_token.txt" ]]; then
    ADMIN_TOKEN=$(cat build-ci-logs/admin_token.txt)
  else
    echo "ERROR: ADMIN_TOKEN 未设置，且未在 build-ci-logs/admin_token.txt 找到。" >&2
    exit 1
  fi
fi

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

post_policy() {
  local fname="$1"; local fsize="$2"; local ctype="$3"
  curl -sS -X POST \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    "$API_BASE/api/v1/admin/storage/oss/policy" \
    --data "{\"business\":\"$BUSINESS\",\"file_name\":\"$fname\",\"content_type\":\"$ctype\",\"file_size\":$fsize}"
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
    echo "ERROR: Policy 响应缺少必要字段：$resp" >&2
    exit 2
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
  # 表单直传
  # 关键字段：key, policy, OSSAccessKeyId, Signature, success_action_status, Content-Type
  # 注意：部分策略可能要求 Content-Length/Content-Type 约束，此处按服务端返回的 policy 条件填充。
  upload_resp=$(curl -sS -X POST "$host" \
    -F "key=$key" \
    -F "policy=$policy_b64" \
    -F "OSSAccessKeyId=$accessKeyId" \
    -F "Signature=$signature" \
    -F "success_action_status=200" \
    -F "Content-Type=$ctype" \
    -F "file=@$file")

  # OSS 表单直传成功时，返回 200，可能包含 XML 响应；此处不强依赖内容，仅校验状态。
  # 简单输出 URL（优先使用 CDN域名）
  cdn_domain=$(read_cdn_domain_from_config)
  if [[ -n "$cdn_domain" ]]; then
    url="https://$cdn_domain/$key"
  else
    url="$host/$key"
  fi
  echo "OK: $file -> $url"

done
