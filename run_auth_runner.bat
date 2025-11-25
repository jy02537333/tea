@echo off
cd /d E:\project\tea
go run scripts\run_auth_tests.go > E:\project\tea\run_auth_tests_stdout.txt 2>&1
exit /b %ERRORLEVEL%
