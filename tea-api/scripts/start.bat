@echo off
echo 茶心阁小程序API服务启动脚本

:: 设置环境变量
set GOOS=windows
set GOARCH=amd64

:: 编译项目
echo 正在编译项目...
go build -o tea-api.exe ./cmd
if %errorlevel% neq 0 (
    echo 编译失败！
    pause
    exit /b 1
)

echo 编译成功！

:: 检查配置文件
if not exist "configs\config.yaml" (
    echo 警告：配置文件 configs\config.yaml 不存在！
    echo 请确保配置文件存在并正确配置数据库连接信息。
    pause
)

:: 启动服务
echo 正在启动茶心阁小程序API服务...
echo 按 Ctrl+C 可以停止服务
echo ====================================
tea-api.exe -config=configs\config.yaml

pause