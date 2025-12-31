@echo off
chcp 65001 >nul
set LANG=zh_CN.UTF-8
set LC_ALL=zh_CN.UTF-8
REM 设置深色主题，提高对比度
set FYNE_THEME=dark
REM 设置缩放比例 1.3 （推荐）
set FYNE_SCALE=1.3
start "" "%~dp0adbmanager.exe"
