# 茶心阁全链路一键部署流程

本指南适用于本地或服务器环境，涵盖 MySQL、Redis、API-Server、Admin-FE 静态页面等全链路部署。

---

## 1. 环境准备

- 安装 [Docker](https://www.docker.com/products/docker-desktop) 和 [Docker Compose](https://docs.docker.com/compose/install/)。
- 确保项目目录结构如下（以 `d:/developTool/work/go/tea` 为例）：
  ```
  tea/
    API-Server/
      Dockerfile
      ...（Go 源码及 configs）
    Admin-FE/
      index.html
      ...（静态前端文件）
    docker-compose.yml
  ```

---

## 2. 配置 `docker-compose.yml`

- 使用如下内容，放在 `tea/` 根目录：

```yaml
version: "3.9"

services:
  mysql:
    image: mysql:8.0
    container_name: tea-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: rootpass
      MYSQL_DATABASE: tea
      MYSQL_USER: teauser
      MYSQL_PASSWORD: teapass
      TZ: Asia/Shanghai
    command:
      - "--character-set-server=utf8mb4"
      - "--collation-server=utf8mb4_general_ci"
      - "--default-authentication-plugin=mysql_native_password"
    ports:
      - "3306:3306"
    volumes:
      - tea-mysql-data:/var/lib/mysql

  redis:
    image: redis:7.0
    container_name: tea-redis
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - tea-redis-data:/data

  tea-api:
    build:
      context: ./tea-api
      dockerfile: Dockerfile
    container_name: tea-api
    restart: unless-stopped
    depends_on:
      - mysql
      - redis
    environment:
      TZ: Asia/Shanghai
      DB_DSN: "teauser:teapass@tcp(mysql:3306)/tea?charset=utf8mb4&parseTime=True&loc=Local"
      REDIS_ADDR: "redis:6379"
    ports:
      - "8080:8080"

  admin-fe:
    image: nginx:1.27-alpine
    container_name: tea-admin-fe
    restart: unless-stopped
    depends_on:
      - api-server
    volumes:
      - ./Admin-FE:/usr/share/nginx/html:ro
    ports:
      - "8081:80"
    environment:
      TZ: Asia/Shanghai

volumes:
  tea-mysql-data:
  tea-redis-data:
```

如需自定义数据库密码、端口、API 配置等，可在 `docker-compose.yml` 里调整。

---

## 3. 构建并启动服务

在 `tea/` 目录下执行：

```powershell
cd d:\developTool\work\go\tea
docker-compose up -d --build
```

- 首次运行会自动拉取 MySQL、Redis、Nginx 镜像，并构建 API-Server。
- 所有服务会以后台模式启动。

---

## 4. 初始化数据库（如有需要）

- 如果 API-Server 启动后自动建表，无需手动操作。
- 如需导入初始 SQL，可进入 MySQL 容器执行：

```powershell
docker exec -it tea-mysql bash
mysql -u teauser -pteapass tea < /path/to/init.sql
```

---

## 5. 访问服务

- 管理后台（Admin-FE）：  
  http://localhost:8081/index.html

- tea-api（RESTful API）：  
  http://localhost:8080/api/v1/...

- MySQL/Redis 可通过容器端口连接，默认账号密码见 compose 文件。

---

## 6. 常用运维命令

- 查看服务状态：
  ```powershell
  docker-compose ps
  ```
- 查看日志：
  ```powershell  
  docker-compose logs -f
  ```
- 停止服务：
  ```powershell
  docker-compose down
  ```
- 重启服务（如有代码更新）：
  ```powershell
  docker-compose up -d --build
  ```

---

如需对接外部小程序、第三方配送等，只需保证 API-Server 的相关端口和配置开放即可。
如有特殊配置（如 Nginx 反向代理、SSL），可在 `admin-fe` 服务的 Nginx 配置中扩展。

如需更详细的定制化部署文档，请补充你的具体环境或特殊需求！

---

## 7. 启用 Claude Sonnet 4.5（可选）

如果你想把系统对话/推理请求切换到 Claude Sonnet 4.5（对所有客户端生效），可以按下面步骤操作：

1) 前提
  - 拥有 Anthropic 提供的 API Key，并确认使用许可和费用预算。
  - 确认你要修改的实例（本地/服务器）能安全注入 Secrets（避免明文写入配置文件）。

2) 配置（推荐通过环境变量 / 容器 Secrets）
  - MODEL_PROVIDER=anthropic
  - MODEL_NAME=claude-sonnet-4.5
  - MODEL_API_KEY=<你的_API_KEY>  # 强烈建议放在 Secrets 中，不要写到 repo
  - 可选：MODEL_API_URL=<自定义 endpoint，如果提供商需要覆盖默认 URL>

3) 在本项目中的快速开关
  - 项目已经在 `configs/config.yaml` 与 `tea-api/configs/config.yaml` 中加入了 `ai:` 配置节，你可以通过把该配置中的 `ai.enabled` 切为 `true` 来开启（仅为说明，生产环境请以环境变量/Secrets 为主）。

4) 重启服务并验证
  - 重启 API 服务；如果使用容器请在 docker-compose 中注入相应环境变量并 `docker-compose up -d --build`，否则在宿主机以系统方式设置环境变量并重启服务进程。
  - 发送 smoke 请求验证模型请求是否成功：

```powershell
$body = @{ prompt = "测试：Hello" } | ConvertTo-Json
Invoke-RestMethod -Uri 'http://localhost:8082/api/v1/model/generate' -Method POST -Body $body -Headers @{ 'Authorization' = "Bearer $env:MODEL_API_KEY"; 'Content-Type' = 'application/json' }
```

5) 监控与回滚
  - 监控延迟、错误率和费用（如 P95、错误率、调用量）。
  - 回滚方法：把 `MODEL_NAME` 或 `MODEL_PROVIDER` 恢复到旧值或清空 Secret，重启服务即可回退。

安全/合规提示：
  - 切勿将 `MODEL_API_KEY` 提交到代码仓库或日志。务必使用 Secret 管理。对敏感 prompt/回复做访问控制和审计。

如果你希望我把一个安全的 adapter（示例 Go 客户端）和自动化 smoke-test 脚本加入仓库以便一键切换并验证，我可以实现：包括一个轻量的 `pkg/modelclient/anthropic.go` 适配器、配置读取、和一个 `scripts/smoke_model.ps1` 验证脚本。请确认是否需要我直接提交这些代码改动。 
