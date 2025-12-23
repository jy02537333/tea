# Sprint C 最小集成（Minimal Integration）

本指南说明如何在本地或 CI 中以“非阻塞”的方式验证 Sprint C 关键接口是否可用，并产出可审计的证据（build-ci-logs/*）。

## 运行最小集成脚本

- 命令：`make run-min-integration`
- 行为：调用 scripts/run_min_integration.sh，自动准备 Token，并尝试访问各个 Sprint C 接口；将返回写入 build-ci-logs/ 下的 JSON/文本。
- 验证：`make verify-sprint-c` 会读取这些证据并生成 build-ci-logs/sprint_c_check_summary.json（CI 可配置为不阻断）。

## SC1：OSS 直传 Policy（Admin）

- 接口：`POST /api/v1/admin/storage/oss/policy`
- 认证：需要 Admin Token（Authorization: Bearer <ADMIN_TOKEN>）
- 请求体示例：

```bash
curl -sS -X POST "$API_BASE/api/v1/admin/storage/oss/policy" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "business":"product_image",
    "file_name":"demo.jpg",
    "content_type":"image/jpeg",
    "file_size":12345
  }'
```

- 期望返回字段：`policy`, `signature`, `accessKeyId`, `expire_at`, `object_key_template`
- 证据文件：`build-ci-logs/get_oss_policy.json`（由脚本自动生成）
- 容错说明：如果后端暂未配置 OSS（或路由暂未就绪），也会将错误 JSON 写入该文件；`verify_sprint_c.sh` 会“软通过”这类占位结果，避免 CI 阻塞。

## 变量约定

- `API_BASE`：后端地址，默认 `http://127.0.0.1:9292`
- `ADMIN_TOKEN` / `USER_TOKEN`：由 `make prepare-tokens` 或脚本自动生成并写入 build-ci-logs/tokens.env

## 故障排查

- 查看 `build-ci-logs/min-integration.log` 获取每步请求的日志。
- 若 `get_oss_policy.json` 不存在或为空，请确认：
  - Token 已生成且为 Admin 角色
  - 后端路由已注册并服务已启动
  - OSS 配置（endpoint/access_key_id/access_key_secret/bucket_name）已在服务端配置
