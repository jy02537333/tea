## 设置 测试环境2 的环境变量（PowerShell）
# 用法：在 PowerShell 中执行： .\scripts\use-test-env2.ps1

# 仓库已禁止使用 SQLite 回退；默认使用 MySQL

# MySQL DSN for 测试环境2
$env:TEA_DSN = 'root:gs963852@tcp(127.0.0.1:3308)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local'

# Redis
$env:REDIS_ADDR = '127.0.0.1:6379'
$env:REDIS_PASS = '123456'

# RabbitMQ
$env:RABBITMQ_ADDR = 'amqp://guest:guest@127.0.0.1:5672/'

Write-Output "已设置 测试环境2 环境变量：TEA_DSN、REDIS_ADDR、RABBITMQ_ADDR（仅当前会话）。"
