# 本地启动指南（Windows + PowerShell）

此文档指导如何在本地（Windows + PowerShell）启动茶心阁项目的各项服务，用于开发/联调。

说明：本说明覆盖两种常用模式：
- 内存版（快速开发、无需 MySQL/Redis）
- 数据库版（连接本机或远程 MySQL + 可选 Redis）
端口约定（项目默认）：
  - Admin-FE（静态页面）: 9294
  - API-Server (数据库版/mock)：9292
  - simple-server / auth-server（内存/演示）: 9292
端口约定（项目默认）：
  - Admin-FE（静态页面）: 8081
  - API-Server (数据库版/mock)：8080
  - simple-server / auth-server（内存/演示）: 8080

---

## 前置准备
go run simple-server.go
- 安装 Go（建议 1.21+）并已配置 GOPATH/GOBIN
  - 检查：`go version`
- 安装 Python 3（用于快速静态文件服务）
  - 检查：`python --version`
- 已安装并运行 MySQL（仅数据库版需要）
  - 检查：本地 MySQL 服务或远程实例
- 可选：Redis（缓存/队列/session）
python -m http.server 9294
配置 JWT Secret（可选，本地调试）
- 若希望覆盖后端本地 mock 的 JWT 签名密钥，可设置环境变量 `TEA_JWT_SECRET`：
  ```powershell
  $env:TEA_JWT_SECRET = "your_local_secret_here"
  ```
  如果不设置，后端会使用内置的开发默认密钥 `dev_secret_change_me`（仅限本地调试，不要用于生产）。
Invoke-RestMethod -Uri "http://localhost:9292/api/v1/health" -Method GET
# 或访问 Admin-FE 页面 http://localhost:9294
---

## 1. 克隆/切换到项目目录

```powershell
cd D:\developTool\work\go\tea
```

---

## 2. 启动内存版 (快速调试，推荐首次联调)

内存版不依赖 MySQL/Redis，适合快速测试前端与基础 API。

```powershell
go run database-server.go
# 会监听端口 8080（项目内约定）
go run simple-server.go
```

在另一个 PowerShell 窗口中启动 Admin-FE（静态文件）：

Invoke-RestMethod -Uri "http://localhost:9292/api/v1/health" -Method GET
Invoke-RestMethod -Uri "http://localhost:9292/admin/users" -Method GET
# 使用 python 的 http.server 快速启动静态服务（端口 8081）
python -m http.server 8081
```

验证：
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET
# 或访问 Admin-FE 页面 http://localhost:8081
go run .
## 3. 启动数据库版（连接 MySQL）

如果你希望使用真实 MySQL 数据库（本机或远程），按下列步骤：

1) 设置环境变量 `TEA_DSN`（示例使用本机 MySQL）
Invoke-RestMethod -Uri "http://localhost:9292/admin/users" -Method GET
```powershell
$env:TEA_DSN = "root:gs963852@tcp(127.0.0.1:3308)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local"
```

或使用生产/测试服务器：

```powershell
# auth-server 默认监听 9292（和 simple-server 互斥）
cd D:\developTool\work\go\tea
go run auth-server.go
2) 启动数据库版后端：

```powershell
# 在项目根运行（会监听 8080）
go run database-server.go
Get-NetTCPConnection -LocalPort 9294,9292,9292 -State Listen | Format-Table -AutoSize
3) 启动 Admin-FE（同上）

```powershell
cd Admin-FE
python -m http.server 8081
```
Start-Process -NoNewWindow -FilePath python -ArgumentList '-m','http.server','9294' -WorkingDirectory 'D:\developTool\work\go\tea\Admin-FE'
4) 验证：
Invoke-RestMethod -Uri "http://localhost:9292/api/v1/health" -Method GET
Invoke-RestMethod -Uri "http://localhost:9294" -Method GET
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET
Invoke-RestMethod -Uri "http://localhost:8080/admin/users" -Method GET
```

备注：如果后端依赖 Redis，请确保 `REDIS_ADDR` 指向可达地址（例如 `127.0.0.1:6379`），并在运行前设置环境变量：

```powershell
$env:REDIS_ADDR = "127.0.0.1:6379"
```

---

## 4. 启动项目自带的 mock API（项目内轻量版）

仓库中可能含有 `API-Server` 的一个轻量 mock，用于兼容 Admin-FE 调试（我们在本机用过）：

```powershell
cd D:\developTool\work\go\tea\API-Server
# 启动 mock 服务（默认监听 8080）。如果想指定端口或使用脚本化启动，建议使用项目根的 `start-dev.ps1`：
# 例如以 8080 启动：
# powershell -ExecutionPolicy Bypass -File ..\start-dev.ps1 -Mode mock -ApiPort 8082
# 或直接运行（若端口空闲）：
go run .
```

验证：
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/admin/users" -Method GET
```

---

## 5. 启动 auth-server（开发登录）

如果需要开发登录（获取管理员令牌）并演示受保护接口：

```powershell
# auth-server 默认监听 8080（和 simple-server 互斥）
cd D:\developTool\work\go\tea
go run auth-server.go
```

使用示例（获取 token）：见 `项目完成总结.md` 中的“开发登录获取管理员令牌”一节。

---

## 6. 启动/停止与端口诊断（常用 PowerShell 命令）

查看进程占用端口：
```powershell
Get-NetTCPConnection -LocalPort 8081,8082,8082 -State Listen | Format-Table -AutoSize
```

根据 PID 停止进程：
```powershell
Stop-Process -Id <PID> -Force
```

查找使用某端口的进程并停止（一行完成）：
```powershell
$pid = (Get-NetTCPConnection -LocalPort 8082 -State Listen).OwningProcess; if($pid){ Stop-Process -Id $pid -Force }
```

---

## 7. Redis 安装/启动建议（Windows）

选项 A：使用 WSL2（推荐）
```bash
# 在 WSL 中
sudo apt update && sudo apt install redis-server -y
sudo service redis-server start
# 验证
redis-cli ping  # 返回 PONG
```

选项 B：使用 Docker（若 Docker 可用）
```powershell
docker run -d --name redis -p 6379:6379 redis:7.0
```

选项 C：Windows 原生 port（不再官方维护），不推荐；建议使用 WSL 或 Docker。

---

## 8. 常见问题与排查

- 后端启动失败：检查 `TEA_DSN` 是否正确，MySQL 是否可达，端口是否被占用。
- Redis 连接失败：检查 `REDIS_ADDR` 与防火墙，或启动 Redis 实例。
- Admin-FE 页面无数据：确认 Admin-FE 的 `API_BASE` 指向正确的 API 地址（在 `index.html` 顶部可配置）。
- 端口已占用：使用 `Get-NetTCPConnection` 找到 PID 并停止旧进程。

---

## 9. 验证脚本示例（一次性在 PowerShell 执行，做快速验证）

```powershell
# 启动内存版 + Admin-FE（示例，手动在不同窗口分别运行更好）
Start-Process -NoNewWindow -FilePath go -ArgumentList 'run','simple-server.go'
Start-Process -NoNewWindow -FilePath python -ArgumentList '-m','http.server','8081' -WorkingDirectory 'D:\developTool\work\go\tea\Admin-FE'
Start-Sleep -Seconds 2
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET
Invoke-RestMethod -Uri "http://localhost:8081" -Method GET
```

---

## 使用 `start-dev.ps1` 自动启动并记录日志

项目提供 `start-dev.ps1`，可在 Windows PowerShell 下一键启动 `Admin-FE` 与 `API-Server`（mock 或完整）。它会创建 `.dev_logs` 目录并把 stdout/stderr 写入日志文件，同时把 PID 与日志路径保存到 `.dev_pids.json`，方便 `stop-dev.ps1` 停止。

示例：在项目根运行（以 8082 启动 API mock）：

```powershell
# 以 mock 模式启动 API-Server（端口 8082），并同时启动 Admin-FE
powershell -ExecutionPolicy Bypass -File .\start-dev.ps1 -Mode mock -ApiPort 8080

# 启动后查看日志
Get-ChildItem .\.dev_logs\*
Get-Content .\.dev_logs\api-server-8082-out.log -Tail 50
Get-Content .\.dev_logs\api-server-8082-err.log -Tail 50
```

若需停止，运行：

```powershell
.\stop-dev.ps1
```

---

## 10. 其他资源

- 若需容器化部署参考：`DEPLOY.md`（项目根，包含 docker-compose 配置示例）
- 项目总体说明：`项目完成总结.md`

---

如需我把 `START_LOCAL.md` 中的某一部分改为脚本化（例如 `start-dev.ps1` / `stop-dev.ps1`），我可以为你生成对应的 PowerShell 脚本并测试。
