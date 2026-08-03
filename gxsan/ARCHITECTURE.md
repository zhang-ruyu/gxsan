# GxSan 架构文档（ARCHITECTURE）

> 版本：步骤3（多页面 + 规划测算落地）后
> 最近更新：2026-07-31

## 1. 总览

GxSan（股息三）是 A 股分红投资分析与决策辅助工具，落地技能 `a-share-dividend-investing` 的方法论。采用 **双模块 + 单一业务真源** 结构：

- **根模块** `github.com/user/gxsan`（CLI）：`/c/gxsan`
- **GUI 模块** `github.com/user/gxsan/gui`（Wails 桌面端）：`/c/gxsan/gxsan`

GUI 通过 `go.mod` 的 `replace github.com/user/gxsan => ../` 直接复用根模块的 `internal/` 业务库，**不在 GUI 内维护重复的 internal 副本**。CLI 与 GUI 共享同一套模型、抓取、测算逻辑，只是入口不同（命令行 vs Vue 前端）。

## 2. 目录结构

```
gxsan/                         # 根模块（CLI）
├── cmd/                       # 命令行入口
├── config/                    # 配置加载/保存（yaml）
├── data/                      # 行情抓取 + 缓存（EastMoneyFetcher, Cache, EnrichStock/EnrichStocks 并发批量）
├── internal/
│   ├── model/                 # 领域模型：Stock/Holding/Account/Config/Portfolio/View DTO
│   ├── config/                # Config Manager（增删持仓/账户/生命周期）
│   ├── data/                  # （同上层 data，内部细分）
│   ├── report/                # 报告生成：analyze/detail/calendar → 文本或 JSON DTO
│   ├── strategy/              # 单只股票买卖信号（DividendStrategy, cost_yield YoC）
│   ├── plan/                  # 【步骤3新增】规划测算领域层
│   ├── fund/                  # 资金/仓位监控（Monitor, BuildPool）
│   └── cli/                   # 命令分发与各子命令实现
└── gxsan/                     # GUI 模块（Wails）
    ├── app.go                 # 薄胶水层：暴露 Wails 绑定方法
    ├── main.go
    ├── frontend/              # Vue3 + Vue Router + Vite
    │   └── src/
    │       ├── views/         # Home/Detail/Portfolio/Pension/Conversion/Calendar/Settings
    │       ├── components/     # Sidebar/StockCard/SignalBadge
    │       ├── router/index.js
    │       └── wailsjs/        # wails generate module 自动生成绑定
    └── wails.json
```

## 3. 分层职责

### model（领域模型，单一真源）
- `stock.go`：`Stock`（行情+分红）、`Holding`（持仓，含 `Account` 字段）、`StockConfig`
- `account.go`：`Account` 类型 + 默认5账户（养老/教育/港美股/娱乐/打新）+ `LifecycleStageName`
- `config.go`：`Config` 增加 `Accounts`、`LifecycleStage`（1启动/2滚雪球/3自由/4收获）
- `portfolio.go`：`InvestPool` + `GroupByAccount()` 按账户分组
- `view.go`：`StockDetail` JSON DTO（history / yield_on_cost / account / 估值区间 / 网格等），对齐 Detail.vue 字段；`PortfolioItem` 增加 `Account`/`YieldOnCost`

### data（抓取与缓存）
- `enrich.go`：`EnrichStock`（单只，内部行情+分红两次 HTTP 并发）、`EnrichStocks(codes, forceRefresh)`（多只批量并发，信号量限流上限 8）
  - 命中 24h 缓存且 `forceRefresh=false` → 直接返回
  - `forceRefresh=true` → 跳过缓存读取（仍写回），用于「同一支股票随时多次刷新」

### report（报告/视图模型）
- `detail.go`：`GenerateStockDetail(code, forceRefresh) (*model.StockDetail, error)` + `FormatStockDetailText(...)` 供 CLI 文本输出
- `report.go`：`GenerateAnalysisReport(forceRefresh bool)`
- 前后端契约：**Detail 返回 `*model.StockDetail` JSON**，前端 `Detail.vue` 直接消费（早期返回纯文本是坏的，已修复）

### strategy（单只信号）
- `cost_yield.go`：`CalculateYoC(avgCost, dps)` 成本股息率
- `DividendStrategy.Analyze`：买卖信号（WATCH/BUY/SELL…）

### plan（规划测算，步骤3新增，独立于 strategy）
- `compound.go`：`CompoundSchedule` 定投复利模拟（月定投×年数×年化，分红100%复投已内含于复利）
- `pension.go`：`PensionPlan` 目标倒推所需本金（年分红÷成本股息率5/6/7%）+ `RetirementNote` 退休提款阶段建议
- `conversion.go`：`CompareStocks` 个股多维对比（含 YoC）+ `RealEstateToEquity` 房产→股权现金流对比

### fund（资金监控）
- `Monitor.BuildPool`：统一计算持仓市值/成本/盈亏/最大仓位

### cli（命令入口）
- `commands.go`：分发 add/list/analyze/detail/pension/compare/yoc/account/config…
- `cmd_*.go`：各子命令；`--refresh` 解析；帮助用 `fmt.Print` 输出（避免「年化%」被当格式符）

## 4. GUI 绑定（app.go）

`app.go` 是薄胶水层，所有逻辑委托给 internal。步骤3新增绑定：
- `RefreshStock(code)` → 强制刷新（绕过缓存），同一支股票可随时多次刷新
- `GetPension(monthly, invest, years, rate)`
- `CompareStocks(codeA, codeB)`、`RealEstateToEquity(principal, reYield, eqYield)`
- `GetAccounts` / `AddAccount` / `RemoveAccount` / `AssignAccount` / `SetLifecycleStage`
- `GetPortfolio` 增加 `account` / `yield_on_cost`
- `GetStockDetail` 返回 JSON（契约对齐）

运行 `wails generate module` 后，`frontend/src/wailsjs/` 自动生成 `RefreshStock`/`GetPension` 等绑定（App.js / App.d.ts）。

## 5. 前端多页面

- 路由（`router/index.js`）：`/`（Home）、`/detail/:code`（Detail）、`/portfolio`（Portfolio）、`/pension`（养老测算）、`/conversion`（资产转换）、`/calendar`、`/settings`
- 侧栏（`Sidebar.vue`）：导航菜单
- **Detail**：刷新按钮调用 `RefreshStock(force)`；展示 YoC、所属账户
- **Portfolio**：按账户分组展示 + YoC 列
- **Pension**：目标月分红/月定投/年数/年化 → 倒推本金 + 定投模拟
- **Conversion**：房产本金/租售比/股息率 → 现金流倍数对比

## 6. 数据流（CLI 与 GUI 共用）

```
用户(CLI参数 / Vue点击)
   │
   ▼
CLI App.Run  /  GUI App 方法
   │  都调用 internal/*
   ▼
report.Reporter / plan.* / fund.Monitor / strategy.*
   │
   ▼
data.EnrichStock (EastMoney 行情 + 24h 缓存, forceRefresh 可绕过)
   │
   ▼
model.* (领域模型)  ←→  config.Manager (yaml 持久化)
```

## 7. 构建与运行

```bash
# CLI
cd /c/gxsan && go build ./... && go run ./cmd

# GUI（需 Wails）
cd /c/gxsan/gxsan && wails dev        # 开发预览
cd /c/gxsan/gxsan && wails build      # 打包（自动 vite build + go build）

# 仅前端构建校验
cd /c/gxsan/gxsan/frontend && npm run build
```
