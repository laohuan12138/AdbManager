@echo off
chcp 65001 >nul
echo 正在编译 ADB 服务管理工具...
echo.

REM 设置 Go 环境变量
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

REM 确保使用 UTF-8 编译
set LANG=zh_CN.UTF-8
set LC_ALL=zh_CN.UTF-8

REM 编译（不隐藏控制台窗口，方便调试）
go build -ldflags="-s -w" -tags="" -o adbmanager.exe .

if %errorlevel% equ 0 (
    echo.
    echo [32m✓ 编译成功！[0m
    echo 生成文件: adbmanager.exe
    echo.
    echo 请使用 "运行.bat" 启动程序以确保中文正确显示
    echo.
) else (
    echo.
    echo [31m✗ 编译失败！[0m
    echo.
    pause
    exit /b 1
)

echo 按任意键退出...
pause >nul
