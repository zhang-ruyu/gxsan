@echo off
REM ============================================================
REM gxsan-sync.bat  -  gxsan multi-machine git sync helper
REM Repo root: C:\soft   (gxsan is a subfolder of it)
REM Usage:
REM   gxsan-sync.bat pull            overwrite local with remote (tracked files)
REM   gxsan-sync.bat pull-clean      overwrite local with remote (also drop untracked)
REM   gxsan-sync.bat push ["msg"]    commit all + push to remote
REM ============================================================
set "REPO=C:\soft"
set "BRANCH=main"

cd /d "%REPO%" || (echo [ERR] cannot enter %REPO% & exit /b 1)

if "%1"=="pull" (
    echo [pull] fetch + reset --hard to origin/%BRANCH% ...
    git fetch origin
    git reset --hard origin/%BRANCH%
    git status
    echo [pull] done. local now matches origin/%BRANCH%
    goto :eof
)

if "%1"=="pull-clean" (
    echo [pull-clean] fetch + reset --hard + clean -fd ...
    git fetch origin
    git reset --hard origin/%BRANCH%
    git clean -fd
    git status
    echo [pull-clean] done.
    goto :eof
)

if "%1"=="push" (
    git add -A
    set "MSG=%~2"
    if "%MSG%"=="" set "MSG=chore: update gxsan"
    git diff --cached --quiet || git commit -m "%MSG%"
    git push origin %BRANCH%
    echo [push] done.
    goto :eof
)

echo Usage: gxsan-sync.bat pull ^| pull-clean ^| push ["commit message"]
exit /b 1
