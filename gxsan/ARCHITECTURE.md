# GxSan 架构文档（ARCHITECTURE）

> 版本：2026-08-05（合并双层目录为单一 module，删除养老测算/资产转换 GUI 页面，新增跟踪推荐）
> 最近更新：2026-08-05

## 1. 总览

GxSan（股息三）是 A 股分红投资分析与决策辅助工具，落地技能 `a-share-dividend-investing` 的方法论。采用 **单一 module + 双入口** 结构：

- **module** `github.com/user/gxsan`（路径 `C:\soft\gxsan`）
- **双入口** `main.go`：`os.Args[1]` 为 CLI 子命令（analyze/detail/pension/compare/yoc/config/version/help）时走命令行；否则启动 Wails GUI
- CLI 与 GUI 共享同一套 `internal/` 业务库，只是入口不同（命令行 vs Vue 前端）

## 2. 目录结构

```
gxsan/                         # 单一 module（CLI + GUI）
├── main.go                    # 双入口：CLI 子命令 / Wails GUI
├── app.go                     # GUI 薄胶水层：Wails 绑定方法
├── go.mod / go.sum            # 含 wails v2 依赖
├── wails.json                 # Wails 配置
├── build.bat                  # 一键构建脚本
├── launch-portable.bat        # 便携启动（WebView2 便携包）
├── build/
│   ├── appicon.ico
│   └── appicon.png
├── winres/                    # Windows 资源文件
├── frontend/                  # Vue3 + Vue Router + Vite
│   ├── dist/                  # vite 构建产物（//go:embed all:frontend/dist）
│   ├── wailsjs/go/main/App.js # 手工维护的前端绑定（wails bindgen 本机跑不了）
│   └── src/
│       ├── views/             # Home/Dashboard/Detail/Portfolio/DividendSummary/Tracking/Calendar/Settings
│       ├── components/        # Sidebar/StockCard/SignalBadge
│       ├── utils/             # signal.js / api.js / market.js
│       └── router/index.js
├── internal/
│   ├── model/                 # 领域模型：Stock/Holding/Config/Portfolio/View DTO/TrackingStock
│   ├── config/                # Config Manager（增删持仓/生命周期/成本修正）
│   ├── data/                  # 行情抓取 + 缓存（EastMoneyFetcher, Cache, EnrichStock/EnrichStocks）
│   ├── report/                # 报告生成：analyze/detail/calendar
│   ├── strategy/              # 单只股票买卖信号（DividendStrategy, cost_yield YoC）
│   ├── plan/                  # 规划测算领域层（compound/pension/conversion）— CLI 仍调用，GUI 已移除对应页面
│   ├── fund/                  # 资金/仓位监控（Monitor, BuildPool, GenerateHoldingRecommendations）
│   └── cli/                   # 命令分发与各子命令实现（含 pension/compare，CLI 独占）
├── config/                    # 配置加载/保存（yaml）
└── data/                      # 行情抓取 + 缓存（与 internal/data 同层）
```

> **注意**：`cmd/` 目录已删除，CLI 子命令内置到 `main.go` 双入口。

## 3. 分层职责

### model（领域模型，单一真源）
- `stock.go`：`Stock`（行情+分红）、`Holding`（持仓，含 TotalCost/OriginalCost）、`StockConfig`
- `account.go`：`LifecycleStageName`（1启动/2滚雪球/3自由/4收获）
- `config.go`：`Config` 含 `LifecycleStage`、`TargetAnnualDividend`
- `portfolio.go`：`InvestPool` 统一计算持仓市值/成本/盈亏
- `view.go`：`StockDetail` JSON DTO、`PortfolioItem`（含 YieldOnCost/OriginalCost）、`ActionAdvice`（含 SuggestedBuyShares/SuggestedSellAmount）、`DividendSummary`
- `tracking.go`：`TrackingStock`/`TrackingCategory` — 31 只推荐标的按 7 大分类硬编码，数据源自 skill 推荐池（截止 2026.7.25）

### data（抓取与缓存）
- `enrich.go`：`EnrichStock`（单只）、`EnrichStocks(codes, forceRefresh)`（批量并发，信号量限流 8）
- `cache.go`：**市场感知动态 TTL**（2026-08-04 修复）— 交易时段（工作日 9:25-15:00 北京时间）3 分钟，非交易时段 8 小时；`forceRefresh=true` 跳过缓存

### report（报告/视图模型）
- `detail.go`：`GenerateStockDetail(code, forceRefresh)` → `*model.StockDetail` JSON
- `report.go`：`GenerateAnalysisReport(forceRefresh)`

### strategy（单只信号）
- `cost_yield.go`：`CalculateYoC(avgCost, dps)` 成本股息率
- `DividendStrategy.Analyze`：买卖信号（BUY/SELL/HOLD/WATCH）

### plan（规划测算 — CLI 独占，GUI 已移除对应页面）
- `compound.go`：`CompoundSchedule` 定投复利模拟
- `pension.go`：`PensionPlan` 目标倒推 + `RetirementNote`
- `conversion.go`：`CompareStocks` + `RealEstateToEquity`
- **说明**：GUI 端已移除养老测算（/pension）和资产转换（/conversion）页面及后端方法（GetPension/CompareStocks/RealEstateToEquity），但 CLI 仍保留 `pension`/`compare` 子命令调用 `plan` 包。Dashboard 的退休金进度条已覆盖养老测算的核心需求。

### fund（资金监控 + 建议引擎）
- `Monitor.BuildPool`：持仓市值/成本/盈亏/最大仓位
- `Monitor.GenerateHoldingRecommendations`：对每只持仓产出 `ActionAdvice`（买/卖/持 + 金额 + 理由 + 约束），复用 `divStrategy.Analyze` + `gridStrategy.Analyze`

### cli（命令入口 — CLI 独占）
- `commands.go`：分发 add/list/analyze/detail/pension/compare/yoc/config/grid/portfolio/fund/calendar/version/help
- `cmd_pension.go`：养老测算 CLI 子命令（GUI 无对应页面）
- `cmd_compare.go`：个股对比 CLI 子命令（GUI 无对应页面）
- `--refresh` 参数：analyze/detail 支持，强制绕过缓存

## 4. GUI 绑定（app.go）

`app.go` 是薄胶水层，所有逻辑委托给 internal。

### 统一错误契约（2026-08-03 重构）
- **查询类接口统一返回 `(string, error)`**：成功返回 JSON 字符串，失败返回 `("", err)`
- 前端 `utils/api.parseJSON(str)` 统一解析

### 绑定清单
| 方法 | 说明 | 缓存策略 |
|------|------|---------|
| `GetAnalysisReport()` | 分析报告 | 走缓存 |
| `GetStockDetail(code)` | 个股详情 JSON | 走缓存 |
| `RefreshStock(code)` | 强制刷新（绕过缓存） | force=true |
| `GetPortfolio()` | 持仓列表（含 OriginalCost/YoC） | 走缓存 |
| `GetActionAdvice()` | 行动提示（买/卖/持+金额+理由） | 走缓存 |
| `GetDashboard()` | 总览（退休金进度条+行动汇总） | 走缓存 |
| `CorrectHoldingCost(code, totalCost)` | 一键修正成本 | 写 config |
| `GetTrackingStocks()` | 跟踪推荐（31 只+实时行情） | 走缓存 |
| `GetDividendSummary()` | 分红按年汇总 | 走缓存 |
| `GetCalendar(days)` | 股利日历 | 文本报告 |
| `GetConfig()` / `SetConfig(key, value)` | 配置管理 | 读写 config |
| `AddHolding` / `RemoveHolding` | 持仓增删 | 写 config |
| `SetLifecycleStage(stage)` | 生命周期阶段 | 写 config |
| `SearchStock(keyword)` | 股票搜索 | 实时 |
| `GetGridStrategy(code)` | 网格策略 | 读 config |
| `GetDataDir()` / `GetConfigFile()` | 工具函数 | — |

> **已移除的 GUI 方法**：`GetPension`、`CompareStocks`、`RealEstateToEquity`（CLI 仍通过 `internal/cli/` 调用 `plan` 包）

### 手工维护前端绑定
本机 Go 工具链较新（go1.26.5），Wails bindgen 用的旧 `golang.org/x/tools@v0.1.12` 会 panic（`unsupported version: 2`）。
新增后端方法后，需手工在 `frontend/wailsjs/go/main/App.js` 补一项：
```js
export function GetXxx(arg1) {
  return window['go']['main']['App']['GetXxx'](arg1);
}
```

## 5. 前端多页面

- 路由（`router/index.js`）：
  - `/` — 分析首页（Home）
  - `/dashboard` — 总览驾驶舱（Dashboard）
  - `/detail/:code` — 个股详情（Detail）
  - `/portfolio` — 持仓管理（Portfolio）
  - `/dividend` — 分红汇总（DividendSummary）
  - `/tracking` — 跟踪推荐（Tracking） **2026-08-05 新增**
  - `/calendar` — 股利日历（Calendar）
  - `/settings` — 系统设置（Settings）

- **已移除的路由**：`/pension`（养老测算）、`/conversion`（资产转换）

### 通用工具（utils/）
- `signal.js`：信号→样式/文案映射（BUY→建议买入、SELL→建议卖出、HOLD→持有、WATCH→关注中）
- `api.js`：`parseJSON` 统一解析
- `market.js`：`isMarketOpen()` / `isPageVisible()` / `refreshReason()` — 智能刷新控制

### 智能刷新
所有页面的 30s 轮询在触发前先判断 `isPageVisible()` 与 `isMarketOpen()`，隐藏页或非交易时段暂停。
**手动点击刷新按钮时**：Detail 调用 `RefreshStock(code)` 强制绕过缓存；其他页面走缓存（靠短 TTL 自动刷新）。

### 术语降噪
前端用户可见文案全部中文：成本股息率（非 YoC）、当前股息率、低估/高估阈值、建议买入/卖出/持有/关注中。
信号键名（BUY/SELL/HOLD/WATCH）仅作为内部标识符，经 signal.js 翻译后显示中文。

## 6. 数据流（CLI 与 GUI 共用）

```
用户(CLI参数 / Vue点击)
   │
   ▼
CLI cli.App.Run  /  GUI app.go 方法
   │  都调用 internal/*
   ▼
report.Reporter / fund.Monitor / strategy.* / plan.*
   │
   ▼
data.EnrichStock (EastMoney 行情 + 动态TTL缓存, forceRefresh 可绕过)
   │
   ▼
model.* (领域模型)  ←→  config.Manager (yaml 持久化)
```

## 7. 构建与运行

> 项目根目录：`C:\soft\gxsan`（单一 module）。
> 所有 git 命令在仓库根 `C:\soft` 执行（gxsan 是其下的子目录）。

### 7.1 构建桌面 exe

```bash
# 前端
cd C:\soft\gxsan\frontend && npm run build   # 产物 -> frontend/dist

# 后端（必须加 -tags production）
cd C:\soft\gxsan && go build -tags production -ldflags="-H windowsgui" -o gxsan.exe .
# 沙箱环境需加 GOSUMDB=off -mod=mod
```

- **必须加 `-tags production`**：Wails v2.5.0 无 tag 时走 `app_default_windows.go` stub 弹错误窗。真正的 build tag 只有 `dev`/`production`/`debug`/`bindings`，没有 `desktop`。
- `-ldflags="-H windowsgui"` 隐藏黑色控制台窗口（GUI 模式）。
- 产物 `gxsan.exe`（约 15.5MB），双击=GUI，命令行带子命令=CLI。
- **永远不要跑 `wails build`**：Go 1.26.5 下 bindgen 必崩（`panic: unsupported version: 2`）。

### 7.2 CLI 构建与校验

```bash
cd C:\soft\gxsan && go build ./...               # 编译全部包
cd C:\soft\gxsan && go test ./internal/plan/...   # 跑规划测算单元测试
cd C:\soft\gxsan\frontend && npm run build        # 前端构建校验
```

### 7.3 分发与 WebView2

- gxsan 桌面端基于 Wails v2，强依赖系统 **WebView2 Runtime**。现代 Win10/11 大多已预装。
- 配置与数据默认在用户主目录 `~/.gxsan/`；`gxsan.exe` 可随意拷贝到其他机器/目录。
