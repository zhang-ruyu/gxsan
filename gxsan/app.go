package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/user/gxsan/internal/config"
	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/fund"
	"github.com/user/gxsan/internal/model"
	"github.com/user/gxsan/internal/report"
	"github.com/user/gxsan/internal/strategy"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 应用程序主结构体（Wails GUI 后端）。
// 作为薄胶水层：所有业务逻辑委托给 internal 下的
// report.Reporter / fund.Monitor / data.EnrichStock，避免与 CLI 重复实现。
type App struct {
	ctx         context.Context
	configMgr   *config.Manager
	fetcher     *data.EastMoneyFetcher
	cache       *data.Cache
	reporter    *report.Reporter
	monitor     *fund.Monitor
	divStrategy *strategy.DividendStrategy
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.configMgr = config.NewManager()
	if err := a.configMgr.Load(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
	}

	a.fetcher = data.NewEastMoneyFetcher()
	a.cache = data.NewCache(a.configMgr.DataPath)
	a.divStrategy = strategy.NewDividendStrategy()
	a.reporter = report.NewReporter(a.configMgr.Config, a.fetcher, a.cache)
	a.monitor = fund.NewMonitor(a.configMgr.Config, a.cache)
}

// toJSON 转换为JSON字符串
func toJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// fetchStocksForCodes 批量并发补全股票行情+分红数据（与 CLI 共用 EnrichStocks）
func (a *App) fetchStocksForCodes(codes []string) map[string]*model.Stock {
	return a.fetcher.EnrichStocks(a.cache, codes, false)
}

// ========== 分析报告 ==========

// GetAnalysisReport 获取完整分析报告
// 统一错误契约：出错时返回 ( "", err )，由 Wails 转为 promise rejection，
// 前端统一在 catch 中处理，不再解析混入 JSON 的 {"error":...}。
func (a *App) GetAnalysisReport() (string, error) {
	reportStr, err := a.reporter.GenerateAnalysisReport(false)
	if err != nil {
		return "", err
	}
	return reportStr, nil
}

// GetStockDetail 获取单只股票详情（读缓存，JSON 契约）
func (a *App) GetStockDetail(code string) (string, error) {
	d, err := a.reporter.GenerateStockDetail(code, false)
	if err != nil {
		return "", err
	}
	return toJSON(d), nil
}

// RefreshStock 强制刷新单只股票（绕过缓存），同一支股票可随时多次刷新
func (a *App) RefreshStock(code string) (string, error) {
	d, err := a.reporter.GenerateStockDetail(code, true)
	if err != nil {
		return "", err
	}
	return toJSON(d), nil
}

// ========== 持仓管理 ==========

// GetPortfolio 获取持仓列表（价格/市值由 fund.Monitor 统一计算）
func (a *App) GetPortfolio() (string, error) {
	codes := make([]string, 0, len(a.configMgr.Config.Portfolio))
	for _, h := range a.configMgr.Config.Portfolio {
		codes = append(codes, h.Code)
	}
	stocks := a.fetchStocksForCodes(codes)
	pool := a.monitor.BuildPool(stocks)

	var items []model.PortfolioItem
	for _, h := range pool.Holdings {
		// 行情是否实时取到：rich map 里有这只股票才算拿到价，否则标记为非实时（UI 显示"行情获取失败"）
		price := h.CurrentPrice
		live := false
		if s, ok := stocks[h.Code]; ok {
			price = s.Price
			live = true
		}
		cost := h.AvgCost * float64(h.Shares)
		profit := h.MarketValue - cost
		profitPct := 0.0
		if cost > 0 {
			profitPct = (profit / cost) * 100
		}
		items = append(items, model.PortfolioItem{
			Code:         h.Code,
			Name:         h.Name,
			Shares:       h.Shares,
			AvgCost:      h.AvgCost,
			OriginalCost: h.OriginalCost,
			Price:        price,
			MarketValue:  h.MarketValue,
			Profit:       profit,
			ProfitPct:    profitPct,
			YieldOnCost:  h.YieldOnCost,
			Live:         live,
		})
	}

	return toJSON(items), nil
}

// AddHolding 添加持仓
func (a *App) AddHolding(code string, name string, shares int, cost float64) error {
	return a.configMgr.AddHolding(code, name, shares, cost)
}

// UpdateHolding 更新持仓
func (a *App) UpdateHolding(code string, shares int) error {
	return a.configMgr.UpdateHolding(code, shares)
}

// RemoveHolding 删除持仓
func (a *App) RemoveHolding(code string) error {
	return a.configMgr.RemoveHolding(code)
}

// ========== 资金管理 ==========

// GetFundInfo 获取资金信息（与 BuildPool 口径一致）
func (a *App) GetFundInfo() (string, error) {
	codes := make([]string, 0, len(a.configMgr.Config.Portfolio))
	for _, h := range a.configMgr.Config.Portfolio {
		codes = append(codes, h.Code)
	}
	stocks := a.fetchStocksForCodes(codes)
	pool := a.monitor.BuildPool(stocks)

	info := model.FundInfo{
		AvailableFund:  pool.AvailableFund,
		TotalMarket:    pool.TotalAsset - pool.AvailableFund,
		TotalAssets:    pool.TotalAsset,
		MaxPosition:    pool.MaxPosition,
		MaxPositionPct: pool.MaxPositionPct,
	}

	return toJSON(info), nil
}

// SetAvailableFund 设置可用资金
func (a *App) SetAvailableFund(amount float64) error {
	a.configMgr.Config.Fund.AvailableFund = amount
	return a.configMgr.Save()
}

// ========== 监控列表 ==========

// GetWatchlist 获取监控列表（并发抓取 + divStrategy 分析）
func (a *App) GetWatchlist() (string, error) {
	codes := make([]string, 0, len(a.configMgr.Config.Watchlist))
	for _, sc := range a.configMgr.Config.Watchlist {
		codes = append(codes, sc.Code)
	}
	stocks := a.fetcher.EnrichStocks(a.cache, codes, false)

	var items []model.WatchlistItem
	for _, sc := range a.configMgr.Config.Watchlist {
		item := model.WatchlistItem{
			Code:        sc.Code,
			Name:        sc.Name,
			TargetYield: sc.TargetYield,
			Signal:      "WATCH",
		}

		if s, ok := stocks[sc.Code]; ok {
			item.Price = s.Price
			item.DividendYield = s.DividendYield
			item.Signal = a.divStrategy.Analyze(s, a.configMgr.Config).Action
		}

		items = append(items, item)
	}

	return toJSON(items), nil
}

// AddWatchlist 添加监控
func (a *App) AddWatchlist(code string, name string, targetYield float64) error {
	return a.configMgr.AddStock(code, name, targetYield)
}

// RemoveWatchlist 删除监控
func (a *App) RemoveWatchlist(code string) error {
	return a.configMgr.RemoveStock(code)
}

// UpdateWatchlist 更新监控
func (a *App) UpdateWatchlist(code string, name string, targetYield float64) error {
	return a.configMgr.UpdateStock(code, name, targetYield)
}

// ========== 搜索 ==========

// SearchStock 搜索股票
func (a *App) SearchStock(keyword string) (string, error) {
	stocks, err := a.fetcher.SearchStock(keyword)
	if err != nil {
		return "", err
	}
	return toJSON(stocks), nil
}

// ========== 配置管理 ==========

// GetConfig 获取配置
func (a *App) GetConfig() (string, error) {
	return toJSON(a.configMgr.Config), nil
}

// SetConfig 设置配置
func (a *App) SetConfig(key string, value string) error {
	switch key {
	case "default_target_yield":
		var v float64
		if _, err := fmt.Sscanf(value, "%f", &v); err == nil {
			a.configMgr.Config.DefaultTargetYield = v
		}
	case "min_dividend_years":
		var v int
		if _, err := fmt.Sscanf(value, "%d", &v); err == nil {
			a.configMgr.Config.MinDividendYears = v
		}
	case "cheap_discount":
		var v float64
		if _, err := fmt.Sscanf(value, "%f", &v); err == nil {
			a.configMgr.Config.CheapDiscount = v
		}
	case "expensive_premium":
		var v float64
		if _, err := fmt.Sscanf(value, "%f", &v); err == nil {
			a.configMgr.Config.ExpensivePremium = v
		}
	case "available_fund":
		var v float64
		if _, err := fmt.Sscanf(value, "%f", &v); err == nil {
			a.configMgr.Config.Fund.AvailableFund = v
		}
	case "max_position_pct":
		var v float64
		if _, err := fmt.Sscanf(value, "%f", &v); err == nil {
			a.configMgr.Config.Fund.MaxPositionPct = v
		}
	case "target_annual_dividend":
		var v float64
		if _, err := fmt.Sscanf(value, "%f", &v); err == nil {
			a.configMgr.Config.TargetAnnualDividend = v
		}
	}
	return a.configMgr.Save()
}

// ========== 日历 ==========

// GetCalendar 获取股利日历
func (a *App) GetCalendar(days int) (string, error) {
	reportStr, err := a.reporter.GenerateCalendar(days)
	if err != nil {
		return "", err
	}
	return reportStr, nil
}

// ========== 工具函数 ==========

// GetDataDir 获取数据目录
func (a *App) GetDataDir() string {
	return a.configMgr.DataPath
}

// ClearCache 清空行情缓存（删除本地所有缓存 json）。
// 当缓存出现脏数据（如历史遗留的 0 价）且短 TTL 尚未过期时，可一键清理强制重抓，
// 不必再让用户去文件系统手动删 json（CLI 路径）。
func (a *App) ClearCache() (string, error) {
	if a.cache == nil {
		return "", fmt.Errorf("缓存未初始化")
	}
	if err := a.cache.Clear(); err != nil {
		return "", err
	}
	return "行情缓存已清空，重新打开持仓/跟踪页或点击刷新即可拉取最新行情", nil
}

// GetConfigFile 获取配置文件路径
func (a *App) GetConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".gxsan", "config.yaml")
}

// GetGridStrategy 获取监控列表股票的网格策略
func (a *App) GetGridStrategy(code string) (string, error) {
	sc := a.configMgr.GetStockConfig(code)
	if sc == nil {
		return "{}", nil
	}
	return toJSON(model.GridStrategyInfo{
		BuyGrids:  sc.GridStrategy.BuyGrids[:],
		SellGrids: sc.GridStrategy.SellGrids[:],
	}), nil
}

// ========== 持仓网格策略（GUI 可编辑） ==========

// GetHoldingGrid 获取某持仓的网格策略 + 当前实时股息率触发的档位与建议动作。
// 持仓的网格档位是用户自定义的（股息率阈值→加仓/减仓金额），与监控列表共用同一套算法。
func (a *App) GetHoldingGrid(code string) (string, error) {
	var holding *model.Holding
	for i := range a.configMgr.Config.Portfolio {
		if a.configMgr.Config.Portfolio[i].Code == code {
			holding = &a.configMgr.Config.Portfolio[i]
			break
		}
	}
	if holding == nil {
		return "{}", fmt.Errorf("持仓 %s 不存在", code)
	}

	stocks := a.fetchStocksForCodes([]string{code})
	yield := 0.0
	if s, ok := stocks[code]; ok {
		yield = s.DividendYield
	}

	gs := strategy.NewGridStrategy()
	sig := gs.AnalyzeYield(yield, holding.GridStrategy.BuyGrids, holding.GridStrategy.SellGrids)

	// 未配置任何网格（买卖档位股息率全为 0）→ 提示先设置，而不是给出无意义的「加仓 0 元」
	configured := false
	for _, g := range holding.GridStrategy.BuyGrids {
		if g.Yield > 0 {
			configured = true
			break
		}
	}
	if !configured {
		for _, g := range holding.GridStrategy.SellGrids {
			if g.Yield > 0 {
				configured = true
				break
			}
		}
	}
	if !configured {
		out := map[string]interface{}{
			"code":          holding.Code,
			"name":          holding.Name,
			"current_yield": yield,
			"buy_grids":     holding.GridStrategy.BuyGrids,
			"sell_grids":    holding.GridStrategy.SellGrids,
			"buy_level":     0,
			"buy_amount":    0,
			"sell_level":    0,
			"sell_percent":  0,
			"action":        "NONE",
			"reason":        "尚未设置网格策略；填写股息率档位并保存后，这里会显示「该加仓/减仓多少」的实时建议",
		}
		return toJSON(out), nil
	}

	out := map[string]interface{}{
		"code":          holding.Code,
		"name":          holding.Name,
		"current_yield": yield,
		"buy_grids":     holding.GridStrategy.BuyGrids,
		"sell_grids":    holding.GridStrategy.SellGrids,
		"buy_level":     sig.BuyLevel,
		"buy_amount":    sig.BuyAmount,
		"sell_level":    sig.SellLevel,
		"sell_percent":  sig.SellPercent,
		"action":        sig.Action,
		"reason":        sig.Reason,
	}
	return toJSON(out), nil
}

// SaveHoldingGrid 保存某持仓的网格策略（5 档买入 + 5 档卖出）。
// buyGrids/sellGrids 为 JSON 数组：[{yield:5.5, amount:5000}, ...]
func (a *App) SaveHoldingGrid(code string, buyGridsJSON string, sellGridsJSON string) error {
	var buyGrids [5]model.GridLevel
	var sellGrids [5]model.GridLevel
	if err := json.Unmarshal([]byte(buyGridsJSON), &buyGrids); err != nil {
		return fmt.Errorf("买入网格格式错误: %w", err)
	}
	if err := json.Unmarshal([]byte(sellGridsJSON), &sellGrids); err != nil {
		return fmt.Errorf("卖出网格格式错误: %w", err)
	}

	for i := range a.configMgr.Config.Portfolio {
		if a.configMgr.Config.Portfolio[i].Code == code {
			a.configMgr.Config.Portfolio[i].GridStrategy = model.GridStrategy{
				BuyGrids:  buyGrids,
				SellGrids: sellGrids,
			}
			return a.configMgr.Save()
		}
	}
	return fmt.Errorf("持仓 %s 不存在", code)
}

// ========== 跟踪推荐 ==========

// GetTrackingStocks 获取跟踪推荐股票列表（skill 推荐池 + 实时行情）
// 返回按主流分类分组的股票列表，每只股票包含 skill 静态数据 + 实时价格/股息率。
func (a *App) GetTrackingStocks() (string, error) {
	codes := model.TrackingAllCodes()
	stocks := a.fetcher.EnrichStocks(a.cache, codes, false)

	// 深拷贝 categories 并填充实时数据
	categories := make([]model.TrackingCategory, len(model.TrackingCategories))
	for i, cat := range model.TrackingCategories {
		stocksCopy := make([]model.TrackingStock, len(cat.Stocks))
		for j, s := range cat.Stocks {
			stocksCopy[j] = s
			if live, ok := stocks[s.Code]; ok {
				stocksCopy[j].Price = live.Price
				stocksCopy[j].CurrentYield = live.DividendYield
				stocksCopy[j].DividendPerShare = live.DividendPerShare
			}
		}
		categories[i] = model.TrackingCategory{
			Name:   cat.Name,
			Stocks: stocksCopy,
		}
	}

	return toJSON(categories), nil
}

// ========== 分红汇总 ==========

// GetDividendSummary 分红按年汇总（养老现金流测算辅助）
// 按自然年汇总各持仓分红金额，便于评估分红现金流与养老本金积累。
func (a *App) GetDividendSummary() (string, error) {
	holdings := a.configMgr.Config.Portfolio
	if len(holdings) == 0 {
		return toJSON(model.DividendSummary{}), nil
	}

	codes := make([]string, 0, len(holdings))
	for _, h := range holdings {
		codes = append(codes, h.Code)
	}
	stocks := a.fetcher.EnrichStocks(a.cache, codes, false)

	byYear := map[int]*model.YearDividend{}
	yearCodes := map[int]map[string]bool{} // 各年参与的持仓去重，避免一年内多次分红重复计数
	totalAnnual := 0.0
	totalMV := 0.0

	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	byStock := make([]model.StockDividend, 0, len(holdings))

	for _, h := range holdings {
		var price, dps float64
		if s, ok := stocks[h.Code]; ok {
			price = s.Price
			dps = s.DividendPerShare
		}

		annual := float64(h.Shares) * dps
		mv := float64(h.Shares) * price

		totalAnnual += annual
		totalMV += mv

		// 该持仓的历年分红记录（优先缓存，EnrichStocks 已落盘）
		var dividends []model.DividendRecord
		if cd, err := a.cache.Get(h.Code); err == nil {
			dividends = cd.Dividends
		}

		// 按年汇总（保留原逻辑）
		for _, d := range dividends {
			var y int
			fmt.Sscanf(d.Date, "%d-", &y)
			if y == 0 {
				continue
			}
			yd := byYear[y]
			if yd == nil {
				yd = &model.YearDividend{Year: y}
				byYear[y] = yd
			}
			yd.TotalDividend += float64(h.Shares) * d.Amount
			if yearCodes[y] == nil {
				yearCodes[y] = map[string]bool{}
			}
			yearCodes[y][h.Code] = true
		}

		// 按股票展现：构造分红事件并按日期倒序
		events := make([]model.StockDividendEvent, 0, len(dividends))
		for _, d := range dividends {
			if d.Date == "" {
				continue
			}
			perShare := d.Amount
			events = append(events, model.StockDividendEvent{
				Date:     d.Date,
				PerShare: perShare,
				Per10:    perShare * 10,
				Per100:   perShare * 100,
			})
		}
		sort.Slice(events, func(i, j int) bool {
			return events[i].Date > events[j].Date
		})

		recent := make([]model.StockDividendEvent, 0)
		older := make([]model.StockDividendEvent, 0)
		for _, e := range events {
			t, err := time.Parse("2006-01-02", e.Date)
			if err != nil {
				older = append(older, e) // 日期无法解析归为更早
				continue
			}
			if t.After(oneYearAgo) || t.Equal(oneYearAgo) {
				recent = append(recent, e)
			} else {
				older = append(older, e)
			}
		}
		// 近一年无记录但有更早记录：取最近一次展示并标记 over_year
		if len(recent) == 0 && len(events) > 0 {
			latest := events[0]
			latest.OverYear = true
			recent = []model.StockDividendEvent{latest}
			older = append([]model.StockDividendEvent{}, events[1:]...)
		}

		byStock = append(byStock, model.StockDividend{
			Code:            h.Code,
			Name:            h.Name,
			Shares:          h.Shares,
			MarketValue:     mv,
			HasData:         len(dividends) > 0,
			RecentDividends: recent,
			OlderDividends:  older,
		})
	}

	for y, codes := range yearCodes {
		if yd, ok := byYear[y]; ok {
			yd.Holdings = len(codes)
		}
	}

	yearList := make([]model.YearDividend, 0, len(byYear))
	for _, v := range byYear {
		yearList = append(yearList, *v)
	}
	sort.Slice(yearList, func(i, j int) bool {
		return yearList[i].Year > yearList[j].Year
	})

	summary := model.DividendSummary{
		ByYear:              yearList,
		ByStock:             byStock,
		TotalAnnualDividend: totalAnnual,
		TotalMarketValue:    totalMV,
		TotalHoldings:       len(holdings),
	}
	return toJSON(summary), nil
}

// ========== 建议引擎 (Phase 1 MVP) ==========

// GetActionAdvice 为持仓组合生成行动提示（买/卖/等 + 金额 + 理由）
// 复用现有 divStrategy + gridStrategy，遍历 config.Portfolio。
func (a *App) GetActionAdvice() (string, error) {
	codes := make([]string, 0, len(a.configMgr.Config.Portfolio))
	for _, h := range a.configMgr.Config.Portfolio {
		codes = append(codes, h.Code)
	}
	stocks := a.fetchStocksForCodes(codes)
	pool := a.monitor.BuildPool(stocks)
	advices := a.monitor.GenerateHoldingRecommendations(pool, stocks)
	return toJSON(advices), nil
}

// GetDashboard 总览：总资产 / 当前年化分红 / 退休金进度条 + 行动提示汇总
func (a *App) GetDashboard() (string, error) {
	holdings := a.configMgr.Config.Portfolio
	codes := make([]string, 0, len(holdings))
	for _, h := range holdings {
		codes = append(codes, h.Code)
	}
	stocks := a.fetchStocksForCodes(codes)
	pool := a.monitor.BuildPool(stocks)
	advices := a.monitor.GenerateHoldingRecommendations(pool, stocks)

	totalMV := 0.0
	for _, h := range pool.Holdings {
		totalMV += h.MarketValue
	}

	totalAnnual := 0.0
	for _, h := range holdings {
		var dps float64
		if s, ok := stocks[h.Code]; ok {
			dps = s.DividendPerShare
		}
		totalAnnual += float64(h.Shares) * dps
	}

	target := a.configMgr.Config.TargetAnnualDividend
	progress := 0.0
	if target > 0 {
		progress = totalAnnual / target * 100
	}

	out := map[string]interface{}{
		"total_assets":            pool.TotalAsset,
		"total_market_value":      totalMV,
		"available_fund":          pool.AvailableFund,
		"current_annual_dividend": totalAnnual,
		"target_annual_dividend":  target,
		"progress_pct":            progress,
		"lifecycle_stage":         model.LifecycleStageName(a.configMgr.Config.LifecycleStage),
		"holdings_count":          len(holdings),
		"advice":                  advices,
		"buy_count":               countAdvice(advices, "BUY"),
		"sell_count":              countAdvice(advices, "SELL"),
		"hold_count":              countAdvice(advices, "HOLD"),
		"watch_count":             countAdvice(advices, "WATCH"),
	}
	return toJSON(out), nil
}

// CorrectHoldingCost 一键修正持仓成本：以真实投入总额推导精确每股成本
func (a *App) CorrectHoldingCost(code string, totalCost float64) error {
	return a.configMgr.CorrectHoldingCost(code, totalCost)
}

// countAdvice 统计某类行动的数量
func countAdvice(advices []model.ActionAdvice, action string) int {
	n := 0
	for _, ad := range advices {
		if ad.Action == action {
			n++
		}
	}
	return n
}

// SetLifecycleStage 设置生命周期阶段（1-4）
func (a *App) SetLifecycleStage(stage int) error {
	return a.configMgr.SetLifecycleStage(stage)
}

// WindowMinimise 最小化窗口（供前端标题栏按钮调用）
func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

// WindowToggleFullscreen 切换全屏（供前端标题栏按钮调用）
func (a *App) WindowToggleFullscreen() {
	if runtime.WindowIsFullscreen(a.ctx) {
		runtime.WindowUnfullscreen(a.ctx)
	} else {
		runtime.WindowFullscreen(a.ctx)
	}
}

// WindowClose 关闭窗口（关闭即退出应用）
func (a *App) WindowClose() {
	runtime.Quit(a.ctx)
}
