# 管理端 OSS 直传策略接口

- 路由：GET /api/v1/admin/storage/oss/policy
- 权限：admin 登录；建议具备财务或运营权限即可
- 用途：用于前端直接将文件上传到阿里云 OSS，无需经过后台中转

## 请求参数
- `dir`：可选，上传目录前缀，若未传则默认 `admin/products/YYYY/MM/DD/`
- `expire`：可选，签名有效期秒数，默认 `1800`（30分钟）

示例：
GET /api/v1/admin/storage/oss/policy?dir=admin/products/2025/12/20/&expire=1200

## 返回示例
```json
{
  "host": "https://<bucket>.<endpoint>",
  "dir": "admin/products/2025/12/20/",
  "policy": "base64-encoded-policy",
  "signature": "base64-encoded-signature",
  "accessKeyId": "AKID...",
  "expire": 1766200000
}
```

## 前端直传参考
- 使用表单 POST 至 `host`，携带以下字段：
  - `key`: 目标对象 Key（需以返回的 `dir` 为前缀）
  - `policy`: 返回的策略
  - `OSSAccessKeyId`: 返回的 `accessKeyId`
  - `signature`: 返回的签名
  - `success_action_status`: 建议设为 `200`
  - `file`: 待上传文件
- 完整示例可参考阿里云 OSS 文档的浏览器表单直传章节。

## 安全说明
- 服务端限制 `starts-with $key dir`，仅允许指定目录前缀；并限制 `content-length-range`，默认 100MB。
- 请结合后端鉴权（admin 登录）使用，避免未授权访问。
