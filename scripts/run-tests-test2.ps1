## 在 测试环境2 下运行 go test（PowerShell helper）
# 作用：临时设置环境变量并运行所有单元测试。
# 用法：在 PowerShell 中执行： .\scripts\run-tests-test2.ps1

Write-Output "Applying 测试环境2 environment..."
. .\scripts\use-test-env2.ps1

Write-Output "Running tests: go test ./..."
go test ./... -v
