@echo off
REM ============================================================
REM  gxsan 本机一键构建脚本（电脑 A / D盘）
REM  用途：用最新源码重新编译出 gxsan.exe
REM  前置：Go 已装到 D:\soft\go；Wails CLI 已装（go install 后位于 %GOPATH%\bin）
REM  用法：双击本文件，或命令行 `gxsan-build.bat`
REM  注意：本脚本只构建，不碰 git（代码请用 git pull 单独同步）
REM ============================================================
setlocal

REM --- 工具链路径（如你 Go 装到别处，改这一行即可）---
set "GO_ROOT=D:\soft\go"
set "GOPATH=%USERPROFILE%\go"

REM --- 把 Go / Wails CLI / gcc(MinGW-w64) 加入 PATH（仅本次会话生效）---
REM     wails build 在 Windows 用 CGO 绑 WebView2，必须 gcc。
REM     下面附带几种常见 gcc 安装路径，装了哪个都能命中。
set "PATH=%GO_ROOT%\bin;%GOPATH%\bin;D:\soft\mingw64\bin;C:\TDM-GCC-64\bin;C:\msys64\mingw64\bin;C:\mingw64\bin;%PATH%"

REM --- 国内代理，加速 go / npm 下载 ---
set "GOPROXY=https://goproxy.cn,direct"
set "GOPRIVATE=github.com/zhang-ruyu/*"
set "NPM_CONFIG_REGISTRY=https://registry.npmmirror.com"

REM --- 进入 GUI 子模块目录 ---
cd /d D:\soft\gxsan-main\gxsan
if errorlevel 1 (
  echo [错误] 找不到 D:\soft\gxsan-main\gxsan，请确认仓库路径。
  pause
  exit /b 1
)

echo ============================================================
echo  gxsan 构建开始（wails build）
echo  源码目录: %CD%
echo ============================================================
call wails build
if errorlevel 1 (
  echo.
  echo [构建失败] 请查看上面的报错信息。
  pause
  exit /b 1
)

echo.
echo ============================================================
echo  [成功] 新 exe 已生成：
echo    D:\soft\gxsan-main\gxsan\build\bin\gxsan.exe
echo  双击它即可打开最新版软件。
echo ============================================================
pause
endlocal
