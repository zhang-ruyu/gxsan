@echo off
REM gxsan 桌面程序构建脚本（阶段一固化）
REM 重要：本脚本用 go build 替代 wails build（Go 1.26.5 下 wails build 的 bindgen 会崩溃）。
REM 运行前请确认：已安装 Node.js（构建前端）与 Go（>=1.21，任意版本均可）。
setlocal

cd /d %~dp0

echo [1/2] 构建前端（npm run build）...
cd frontend
call npm run build
if errorlevel 1 (
    echo 前端构建失败，已中止。
    exit /b 1
)
cd ..

echo [2/2] 构建桌面 exe（go build + Wails 生产标签，禁止使用 wails build）...
REM Wails v2 运行时要求生产构建必须带 desktop,production 标签；
REM -H windowsgui 让 exe 以 GUI 方式启动（不显示黑色控制台窗口）。
go build -tags desktop,production -ldflags="-H windowsgui" -o gxsan.exe .
if errorlevel 1 (
    echo go build 失败，已中止。
    exit /b 1
)

echo 构建完成，产物位于：gxsan\gxsan\gxsan.exe
echo 提示：标准分发只需这一个 gxsan.exe；目标机若缺 WebView2 见 README 的「分发与 WebView2」。
endlocal
