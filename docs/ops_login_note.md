登录与 token 获取（运维笔记，**不包含真实敏感秘钥/token**）

目的：记录如何在本地或运维环境复现 Admin-FE 的登录流程并获取用于临时调试的 token（仅用于排障/本地验证）。请勿在文档中保存长期凭证或将 token 上传到公共仓库。

1) 快速步骤（命令示例）

- 启动后端（在 API-Server 目录）：
```powershell
cd d:\developTool\work\go\tea\API-Server
# 默认将后端在本地运行于 8080（可根据需要替换）
$env:PORT='8080'
# 启动开发 mock 服务（以后台进程方式运行）
```

- 启动 Admin-FE 静态服务（在 Admin-FE 目录）：
```powershell
cd d:\developTool\work\go\tea\Admin-FE
# 简单静态服务器（仅用于本地调试）
python -m http.server 8000
 启动后端（在 API-Server 目录）：
```

- 获取 token（PowerShell，不将 token 写入仓库）：
```powershell
$login = Invoke-RestMethod -Uri 'http://localhost:8080/auth/login' -Method POST -Body (ConvertTo-Json @{ username='admin'; password='pass'}) -ContentType 'application/json'
$tok = $login.data.token
Write-Host 'Token (copy to secure place if needed):' $tok
```
- 使用 token 调用受保护接口（示例）：
```powershell
Invoke-RestMethod -Uri 'http://localhost:8080/admin/menus' -Method GET -Headers @{ Authorization = "Bearer $tok" }
```

 获取 token（PowerShell，不将 token 写入仓库）：
2) 使用自动化脚本（更快复现）

- 上面 repo 已包含一个 Playwright 脚本 `tools/browser-login/login-test.js`，可自动打开 Admin-FE、执行登录并把测试报告写入 `tools/browser-login/report-*.json`。

运行示例：
```powershell
npm install
npx playwright install
$env:ADMIN_FE_URL='http://localhost:8000'
 使用 token 调用受保护接口（示例）：
$env:TEST_USER='admin'
$env:TEST_PASS='pass'
node login-test.js
# 执行后会在该目录生成 report-<timestamp>.json，包含登录时读取到的 localStorage token（仅用于本地/受控测试）
```
3) 安全注意事项
- 永远不要把运行时 token 或 API Key 上传到公开仓库；如果需要分享用于排障的 token，请使用受控渠道（内部临时凭证、加密聊天或秘密管理系统）。
- 在生产环境中使用更短生命周期的凭证和审计日志，并把 `MODEL_API_KEY` / `TEA_JWT_SECRET` 等敏感信息放入 Secret 管理系统（K8s Secret / Cloud Secret / Windows Secret Manager）。
 如果浏览器登录后未写入 `localStorage.tea_admin_token`：

4) 常见问题排查
- 如果浏览器登录后未写入 `localStorage.tea_admin_token`：
  - 检查浏览器 DevTools 的 Network 面板，看是否有 `POST /auth/login` 返回 200；
 如果 `Invoke-RestMethod` 报错连接拒绝：确认 API 服务已在 8080 上监听（`Get-NetTCPConnection -LocalPort 8080 -State Listen`）。
  - 确保 `tea-api` 在 `:8080` 正常运行，并返回 CORS 头（`Access-Control-Allow-Origin: *`）用于本地调试；

- 如果 `Invoke-RestMethod` 报错连接拒绝：确认 API 服务已在 8080 上监听（`Get-NetTCPConnection -LocalPort 8080 -State Listen`）。

---

此文档仅用于团队内部快速复现与排障；对于长期凭证管理请参见公司安全/运维规范。