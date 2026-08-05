@echo off
REM gxsan 便携启动模板（仅当目标机缺系统 WebView2 Runtime 时使用）
REM
REM 标准分发只发 gxsan.exe（现代 Win10/11 大多已自带 WebView2，直接双击即可）。
REM 仅当目标机没有 WebView2、又无法联网安装 Runtime 时，才用本模板走「Fixed Version」离线方案：
REM
REM   1. 打开 https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/
REM      下载「Fixed Version」（常青版依赖在线安装，不适合离线）；选与目标机架构一致（多为 x64）的版本。
REM   2. 解压后得到一个含 msedgewebview2.exe 的文件夹，把该文件夹改名为 webview2
REM      并放在本脚本同目录（也就是 gxsan.exe 旁边）。
REM   3. 双击本脚本启动 gxsan。
REM
REM 原理：设置 WEBVIEW2_BROWSER_EXECUTABLE_FOLDER 环境变量，让 Wails 使用此文件夹里的 WebView2，
REM 而不依赖系统全局 Runtime。
setlocal

cd /d %~dp0

if not exist "webview2\msedgewebview2.exe" (
    echo [错误] 未找到 webview2\msedgewebview2.exe
    echo 请先从 Microsoft 官网下载 WebView2 Fixed Version，解压到本目录下的 webview2\ 文件夹。
    echo 下载地址: https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/
    pause
    exit /b 1
)

set "WEBVIEW2_BROWSER_EXECUTABLE_FOLDER=%~dp0webview2"
echo 使用便携 WebView2 启动 gxsan...
start "" "%~dp0gxsan.exe"

endlocal
