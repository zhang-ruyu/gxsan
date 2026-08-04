# gxsan 优化落地路线图（待确认版）

> 来源：2026-08-04 用户结构化 code review + 双方对齐的优先级
> 状态：**方案文档，尚未执行**。确认后按阶段逐项落地并提交。
> 原则（已对齐）：前后分离、算法在 `internal/` 纯 Go、分发干净无 Python/AKShare。

---

## 总体排期

| 阶段 | 目标 | 包含项 |
|------|------|--------|
| **阶段一** | 能交到别人手里 | 构建脚本 + README、WebView2 固定版本打包、plan 包单元测试 |
| **阶段二** | 体验过关 | 便携模式、全局错误 Toast、节假日日历 |
| **阶段三** | 工程化 | slog 日志、Pinia、自动更新、CI 交叉编译 |

---

## 阶段一：能交出去（P0）

### 1.1 构建流程固化（P0-1）
**问题**：`wails build` 在 Go 1.26.5 下 bindgen 崩溃；每次重打要记两条命令。
**结论（微调）**：不必降级 Go——我们用 `go build` 替代，它在 1.26.5 正常。固化"正确命令"即可。

新增 `C:\soft\gxsan\gxsan\build.bat`：
```bat
@echo off
cd /d %~dp0
echo [1/2] 构建前端...
cd frontend
call npm run build
if errorlevel 1 exit /b 1
echo [2/2] 构建桌面 exe（用 go build，不要用 wails build）...
cd ..
go build -o gxsan.exe .
if errorlevel 1 exit /b 1
echo 完成 -> gxsan\gxsan\gxsan.exe
```
- **README（或 ARCHITECTURE.md）新增「构建」一节**，明确写：
  - 改动前端后先 `npm run build`，再 `go build -o gxsan.exe .`
  - **永远不要跑 `wails build`**（Go 1.26.5 下必崩；已用 `go build` 替代）
  - Go 版本：任意 ≥1.21 均可（go build 不卡版本）；`go.mod` 现有 `go 1.21` 仅作声明
  - 若新增后端方法：手工在 `frontend/wailsjs/go/main/App.js` 补一项，或兼容工具链下 `wails dev` 自动重生成
- 涉及文件：`gxsan/gxsan/build.bat`（新）、`gxsan/ARCHITECTURE.md`（补「构建」节）

### 1.2 WebView2 分发策略（P0-2，已决策：标准分发**不捆绑** Fixed Version）
**决策**：Fixed Version ~150MB 太重，标准分发物**只发 `gxsan.exe`**。现代 Win10/11 大多已自带系统 WebView2，直接双击即可。
**按需方案**（仅当目标机缺 WebView2 时）：
- ① 安装 [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)；或
- ② 单独获取便携包：`webview2/<版本>/` 文件夹 + `launch-portable.bat`（设 `WEBVIEW2_BROWSER_EXECUTABLE_FOLDER` 指向它）启动。
提供 `launch-portable.bat` **模板**（不含 150MB 文件夹），需要时填入版本号即用。
- 涉及文件：`gxsan/gxsan/launch-portable.bat`（新，模板）、`.gitignore`（忽略 `webview2/`）

### 1.3 plan 包单元测试（原 P2-8，提升到 P0）
**问题**：养老测算 / 资产转换是核心金融计算，零测试；刚改过小数位，精度 bug 用户难发现。
**方案**：给 `internal/plan` 写 table-driven tests，覆盖边界。
- 文件：`internal/plan/pension_test.go`、`compound_test.go`、`compare_test.go`、`realestate_test.go`
- 用例：股息率=0、负收益率、超长年限、inf/NaN 防护、3 位小数一致性
- 验证：`cd C:\soft\gxsan && go test ./internal/plan/...`
- 涉及文件：`internal/plan/*_test.go`（新）

---

## 阶段二：体验过关（P1）

### 2.1 便携模式（P1-3）
**问题**：配置在 `~/.gxsan/`，exe 拷到别的电脑持仓不带过去。
**方案**：检测 exe 同目录是否有 `config.yaml`，有则整目录当数据根。
- 文件：`internal/config/config.go`
```go
exeDir, _ := filepath.Abs(filepath.Dir(os.Executable()))
if _, err := os.Stat(filepath.Join(exeDir, "config.yaml")); err == nil {
    configDir = exeDir          // 便携：配置+缓存放 exe 旁边
} else {
    configDir = filepath.Join(home, ".gxsan")  // 默认：用户主目录
}
```
- 涉及文件：`internal/config/config.go`

### 2.2 全局错误 Toast（P1-4）
**问题**：`alert()` 在桌面应用里很违和。
**方案**：新增 `Toast` 组件 + 轻量事件总线，替换 `alert()`。
- 文件：`frontend/src/components/Toast.vue`（新）、`frontend/src/utils/toast.js`（新，用 mitt 或 reactive）
- 改造：`Conversion.vue`（2 处 alert）、`Settings.vue` 中的 alert → `toast.error(...)`
- 涉及文件：Toast.vue、toast.js、Conversion.vue、Settings.vue

### 2.3 节假日日历（P1-5，已决策：运行时联网抓取，不硬编码）
**问题**：智能刷新只判交易时段，没判法定节假日（国庆/春节等休市）。
**方案**：运行时联网抓取官方休市日历，不硬编码。
- 后端新增 `GetTradingHolidays(year)` 接口：拉取上交所/深交所或第三方节假日 API，缓存到 `data/`（按年）；前端 `utils/market.js` 的 `isMarketOpen` 增加 `!isTradingHoliday(now)` 判断。
- 兜底：抓取失败则用「周末 + 已知固定长假」近似，并 `console.warn` 提示；缓存按年刷新，避免每次联网。
- 涉及文件：`internal/...`（新接口）、`frontend/src/utils/market.js`

---

## 阶段三：工程化（P2）

| 项 | 内容 | 文件/动作 |
|----|------|----------|
| **slog 日志 (P2-9)** | `fmt.Println` → `log/slog`，支持 INFO/WARN/ERROR，可用 env/flag 调级别 | `app.go`/`report.go`/`enrich.go` 替换 |
| **Pinia (P2-6)** | 引入 Pinia，把股票/持仓/配置抽到 store，页面只展示（大重构，谨慎） | 新增 `stores/`，改 8 个页面 |
| **自动更新 (P2-7)** | 短期：「检查更新」按钮比对 GitHub Releases 版本号；长期：`go-selfupdate` | 新增 `GetLatestRelease` 后端方法 + 设置页按钮 |
| **CI 交叉编译 (P2-10)** | GitHub Actions 矩阵 win/mac/linux；`internal/` 已跨平台，GUI 层需各平台 webview | `.github/workflows/build.yml`（注意：Mac/Linux 的 GUI 也依赖各自 webview 运行时） |

---

## 技术债闭环（认可项）

| 问题 | 根因 | 处理 |
|------|------|------|
| `wails build` 崩溃 | Go 1.26.5 导出格式旧 bindgen 不认 | 已用 `go build` 替代（见 1.1） |
| `gui.exe` vs `gxsan.exe` 混淆 | 两个 exe 共存 | 已删 `gui.exe`，README 注明只用 `gxsan.exe` |
| 前端 dist 清空失败 | Vite 的 safe-delete shim | `NODE_OPTIONS="" npm run build`；CI 固化此环境 |
| `gxsan-sync.bat push` bug | `git diff --cached --quiet` 返回值理解反 | 已修复（2026-08-03） |

---

## 开放决策（已确认）
1. **WebView2 Fixed Version**：标准分发**不捆绑**（太重）。仅按需提供便携包。
2. **Go 版本**：**不锁** go.mod（维持现有 `go 1.21` 声明）。靠 README 写清 `go build` 命令即可；锁版本在 `go build` 替代方案下无意义。
3. **节假日数据**：**运行时联网抓取**官方休市日历（不硬编码），加缓存与失败兜底。

---

## 预计工作量（粗估）
- 阶段一：约 0.5~1 天（含 WebView2 下载、测试编写与跑通）
- 阶段二：约 1 天
- 阶段三：slog 0.5 天；其余按需，Pinia/CI 各 1~2 天

> 确认后我从阶段一开始，每完成一项即提交并（按需）推送到 GitHub。
