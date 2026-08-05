@echo off
REM gxsan 桌面程序构建脚本（阶段一固化）
REM 重要：本脚本用 go build 替代 wails build（Go 1.26.5 下 wails build 的 bindgen 会崩溃）。
REM 运行前请确认：已安装 Node.js（构建前端）与 Go（>=1.21，任意版本均可）。
setlocal

cd /d %~dp0

echo [1/3] 生成图标资源（go-winres 把 winres/appicon.png 嵌入 exe，解决任务栏图标糊/无图标）...
go-winres make --in winres/winres.json --arch amd64 --out appicon
if errorlevel 1 (
    echo 警告：go-winres 未安装或失败，将不嵌入图标（exe 用默认图标）。
)

echo [2/3] 构建前端（npm run build）...
cd frontend
call npm run build
if errorlevel 1 (
    echo 前端构建失败，已中止。
    exit /b 1
)
cd ..

echo [3/3] 构建桌面 exe（go build + Wails 生产标签，禁止使用 wails build）...
REM Wails v2 运行时要求生产构建必须带 production 标签；
REM -H windowsgui 让 exe 以 GUI 方式启动（不显示黑色控制台窗口）。
go build -tags production -ldflags="-H windowsgui" -o gxsan.exe .
if errorlevel 1 (
    echo go build 失败，已中止。
    exit /b 1
)

echo 构建完成，产物位于：gxsan\gxsan\gxsan.exe
echo 提示：标准分发只需这一个 gxsan.exe；目标机若缺 WebView2 见 README 的「分发与 WebView2」。
endlocal
