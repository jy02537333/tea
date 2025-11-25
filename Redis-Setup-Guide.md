# Redis 安装和配置指南

## 方法1：使用Windows Redis 二进制文件

### 下载和安装
1. 访问 https://github.com/microsoftarchive/redis/releases
2. 下载最新的 Redis-x64-xxx.zip
3. 解压到 C:\Redis\ 目录

### 配置文件修改
编辑 `C:\Redis\redis.windows.conf` 文件：
```
# 设置密码
requirepass 123456

# 绑定地址
bind 127.0.0.1

# 端口
port 6379

# 后台运行
daemonize no
```

### 启动Redis服务器
```powershell
# 进入Redis目录
cd C:\Redis

# 启动Redis服务器
.\redis-server.exe redis.windows.conf
```

### 测试连接
```powershell
# 使用Redis客户端测试
.\redis-cli.exe -h 10.8.0.14 -p 6379 -a 123456

# 测试命令
10.8.0.14:6379> ping
PONG
```

## 方法2：使用Docker

### 启动Redis容器
```powershell
# 拉取Redis镜像
docker pull redis:latest

# 启动Redis容器，设置密码
docker run --name tea-redis -p 6379:6379 -d redis:latest redis-server --requirepass 123456

# 测试连接
docker exec -it tea-redis redis-cli -a 123456 ping
```

## 方法3：Windows Subsystem for Linux (WSL)

如果你安装了WSL，可以在Ubuntu中安装Redis：
```bash
# 更新包列表
sudo apt update

# 安装Redis
sudo apt install redis-server

# 编辑配置文件
sudo nano /etc/redis/redis.conf

# 修改以下配置：
# bind 127.0.0.1
# requirepass 123456

# 启动Redis服务
sudo systemctl start redis-server
sudo systemctl enable redis-server

# 测试连接
redis-cli -h 127.0.0.1 -p 6379 -a 123456 ping
```

## 推荐方案

对于开发环境，推荐使用 **方法1（Windows二进制文件）** 或 **方法2（Docker）**。

### 快速启动脚本 (PowerShell)

创建 `start-redis.ps1` 文件：
```powershell
# Redis启动脚本
$redisPath = "C:\Redis"
$configFile = "$redisPath\redis.windows.conf"

if (Test-Path $redisPath) {
    Write-Host "启动Redis服务器..." -ForegroundColor Green
    Set-Location $redisPath
    .\redis-server.exe $configFile
} else {
    Write-Host "Redis未安装，请先下载并解压到 C:\Redis\" -ForegroundColor Red
    Write-Host "下载地址: https://github.com/microsoftarchive/redis/releases" -ForegroundColor Yellow
}
```

## 验证Redis连接

安装启动Redis后，重新启动茶心阁API服务器：
```powershell
cd d:\developTool\work\go\tea\tea-api
go build -o tea-server-final.exe ./cmd/main.go
.\tea-server-final.exe
```

看到以下输出说明Redis连接成功：
```
正在连接Redis: 127.0.0.1:6379
Redis连接成功!
```