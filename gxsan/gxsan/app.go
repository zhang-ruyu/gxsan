package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/user/gxsan/internal/config"
	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/fund"
	"github.com/user/gxsan/internal/model"
	"github.com/user/gxsan/internal/plan"
	"github.com/user/gxsan/internal/report"
	"github.com/user/gxsan/internal/strategy"
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
		cost := h.AvgCost * float64(h.Shares)
		profit := h.MarketValue - cost
		profitPct := 0.0
		if cost > 0 {
			profitPct = (profit / cost) * 100
		}
		items = append(items, model.PortfolioItem{
			Code:        h.Code,
			Name:        h.Name,
			Shares:      h.Shares,
			AvgCost:     h.AvgCost,
			Price:       h.CurrentPrice,
			MarketValue: h.MarketValue,
			Profit:      profit,
			ProfitPct:   profitPct,
			YieldOnCost: h.YieldOnCost,
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

// GetConfigFile 获取配置文件路径
func (a *App) GetConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".gxsan", "config.yaml")
}

// GetGridStrategy 获取网格策略
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

// ========== 成本股息率(YoC) 辅助 ==========

// holdingYoC 计算某持仓的成本股息率（无持仓返回0）
func holdingYoC(cfg *model.Config, code string, s *model.Stock) float64 {
	for _, h := range cfg.Portfolio {
		if h.Code == code && h.AvgCost > 0 {
			return data.CalculateDividendYield(h.AvgCost, s.DividendPerShare)
		}
	}
	return 0
}

// ========== 养老测算 ==========

// GetPension 养老现金流测算（目标倒推 + 定投复利模拟 + 退休提款）
// 参数: monthly=目标月分红, invest=月定投, years=年数, rate=年化收益率(%)
func (a *App) GetPension(monthly, invest float64, years int, rate float64) (string, error) {
	r := rate / 100
	out := map[string]interface{}{
		"monthly":          monthly,
		"invest":           invest,
		"years":            years,
		"rate":             r,
		"lifecycle_stage":  model.LifecycleStageName(a.configMgr.Config.LifecycleStage),
		"plan":             plan.PensionPlan(monthly, []float64{0.05, 0.06, 0.07}),
		"schedule":         plan.CompoundSchedule(invest, years, r),
		"retirement_note":  plan.RetirementNote(a.configMgr.Config.LifecycleStage),
	}
	return toJSON(out), nil
}

// ========== 资产转换 / 个股对比 ==========

// CompareStocks 两只股票多维对比
func (a *App) CompareStocks(codeA, codeB string) (string, error) {
	stocks := a.fetcher.EnrichStocks(a.cache, []string{codeA, codeB}, false)
	sa, okA := stocks[codeA]
	sb, okB := stocks[codeB]
	if !okA || !okB {
		return "", fmt.Errorf("获取股票失败 A=%s B=%s", codeA, codeB)
	}
	r := plan.CompareStocks(sa, sb,
		holdingYoC(a.configMgr.Config, codeA, sa),
		holdingYoC(a.configMgr.Config, codeB, sb))
	return toJSON(r), nil
}

// RealEstateToEquity 房产→红利股权现金流对比
// 参数: principal=本金, reYield=租售比(%), eqYield=红利股息率(%)
func (a *App) RealEstateToEquity(principal, reYield, eqYield float64) (string, error) {
	return toJSON(plan.RealEstateToEquity(principal, reYield/100, eqYield/100)), nil
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

		// 按年汇总：取该持仓的历年分红记录（优先缓存，EnrichStocks 已落盘）
		var dividends []model.DividendRecord
		if cd, err := a.cache.Get(h.Code); err == nil {
			dividends = cd.Dividends
		}
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
		TotalAnnualDividend: totalAnnual,
		TotalMarketValue:    totalMV,
		TotalHoldings:       len(holdings),
	}
	return toJSON(summary), nil
}

// SetLifecycleStage 设置生命周期阶段（1-4）
func (a *App) SetLifecycleStage(stage int) error {
	return a.configMgr.SetLifecycleStage(stage)
}
