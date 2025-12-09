<#
Start script: launch backend and two frontends in separate PowerShell windows

Usage (from project root):
  powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\start-all-test2.ps1

This script will open three new PowerShell windows and run:
  - tea-api (go run .) at `E:\project\tea\tea-api`
  - Admin-FE (python -m http.server 8081) at `E:\project\tea\Admin-FE`
  - WX-FE (python -m http.server 8082) at `E:\project\tea\WX-FE`

Edit the variables section to change paths or ports.
#>

## --- 可调整的变量 ---
$TeaApiPath = "E:\project\tea\tea-api"
$AdminFePath = "E:\project\tea\Admin-FE"
$WxFePath = "E:\project\tea\WX-FE"
$AdminPort = 8081
$WxPort = 8082

# 测试环境2 的环境变量（按需修改）
$TeaDsn = 'root:gs963852@tcp(127.0.0.1:3308)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local'
$RedisHost = '127.0.0.1'
$RedisPort = 6379
$RedisPass = ''
$RedisAddr = "$RedisHost:$RedisPort"
$RabbitAddr = 'amqp://guest:guest@127.0.0.1:5672/'

Write-Output "Preparing to use Test Environment 2 and start services..."

## 后端启动命令串（在新窗口中运行，以便保留日志）
$apiCmd = @'
$env:TEA_USE_SQLITE = '0'
$env:TEA_DSN = '{0}'
# Also set TEA_DATABASE_* vars because the app's config loader prefers those
$env:TEA_DATABASE_HOST = '{4}'
$env:TEA_DATABASE_PORT = '{5}'
$env:TEA_DATABASE_USERNAME = '{6}'
$env:TEA_DATABASE_PASSWORD = '{7}'
$env:TEA_DATABASE_DBNAME = '{8}'
# Ensure TEA_REDIS_* are set so viper can pick them up
$env:TEA_REDIS_HOST = '{10}'
$env:TEA_REDIS_PORT = '{11}'
$env:TEA_REDIS_PASSWORD = '{12}'
$env:REDIS_ADDR = '{1}'
$env:REDIS_PASS = '{2}'
$env:RABBITMQ_ADDR = '{3}'
Set-Location -Path '{9}'
Write-Output 'Starting tea-api (go run .) — console will remain open for logs.'
go run .
'@ -f $TeaDsn, $RedisAddr, $RedisPass, $RabbitAddr, '127.0.0.1', 3306, 'root', 'gs963852', 'tea_shop', $TeaApiPath, $RedisHost, $RedisPort, $RedisPass

## Admin-FE 静态服务器命令
$adminCmd = @'
Set-Location -Path '{0}'
Write-Output '启动 Admin-FE 静态服务器: http://localhost:{1}'
python -m http.server {1}
'@ -f $AdminFePath, $AdminPort

## WX-FE 静态服务器命令
$wxCmd = @'
Set-Location -Path '{0}'
Write-Output '启动 WX-FE 静态服务器: http://localhost:{1}'
python -m http.server {1}
'@ -f $WxFePath, $WxPort

function Start-NewWindow($command, $title) {
  Start-Process -FilePath "powershell" -ArgumentList @('-NoExit','-NoProfile','-Command',$command) -WindowStyle Normal -WorkingDirectory $PWD | Out-Null
  Start-Sleep -Milliseconds 300
}

try {
    Write-Output "Opening backend window..."
    Start-NewWindow $apiCmd "tea-api"

    Write-Output "Opening Admin-FE window..."
    Start-NewWindow $adminCmd "Admin-FE"

    Write-Output "Opening WX-FE window..."
    Start-NewWindow $wxCmd "WX-FE"

    Write-Output "Launched three services in separate windows.`nNote: if you see permission or execution policy errors, run this script as Administrator or use -ExecutionPolicy Bypass."
  } catch {
    Write-Error "Error while starting services: $_"
  }
