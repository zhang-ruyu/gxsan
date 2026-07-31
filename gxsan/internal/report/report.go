package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/fund"
	"github.com/user/gxsan/internal/model"
	"github.com/user/gxsan/internal/strategy"
)

// Reporter 报告生成器：负责编排分析流程并渲染文本报告。
// 数据补全（行情+分红+缓存）已下沉到 data.EnrichStock，
// 单只详情见 detail.go，股利日历见 calendar.go。
type Reporter struct {
	config       *model.Config
	fetcher      *data.EastMoneyFetcher
	cache        *data.Cache
	divStrategy  *strategy.DividendStrategy
	gridStrategy *strategy.GridStrategy
	monitor      *fund.Monitor
}

// NewReporter 创建报告生成器
func NewReporter(config *model.Config, fetcher *data.EastMoneyFetcher, cache *data.Cache) *Reporter {
	return &Reporter{
		config:       config,
		fetcher:      fetcher,
		cache:        cache,
		divStrategy:  strategy.NewDividendStrategy(),
		gridStrategy: strategy.NewGridStrategy(),
		monitor:      fund.NewMonitor(config, cache),
	}
}

// GenerateAnalysisReport 生成分析报告
// forceRefresh 为 true 时跳过缓存，重新抓取所有监控股票数据。
func (r *Reporter) GenerateAnalysisReport(forceRefresh bool) (string, error) {
	var result strings.Builder

	// 获取所有股票数据
	stocks := make(map[string]*model.Stock)
	for _, stockConfig := range r.config.Watchlist {
		stock, err := r.fetcher.EnrichStock(r.cache, stockConfig.Code, forceRefresh)
		if err != nil {
			continue
		}
		stocks[stockConfig.Code] = stock
	}

	// 构建投资池
	pool := r.monitor.BuildPool(stocks)

	// 报告头部
	result.WriteString(r.generateHeader())
	result.WriteString(r.generateConfigSummary())
	result.WriteString("\n")

	// 分析每只股票
	buyCount := 0
	holdCount := 0
	sellCount := 0
	watchCount := 0

	for _, stockConfig := range r.config.Watchlist {
		stock, ok := stocks[stockConfig.Code]
		if !ok {
			continue
		}

		signal := r.divStrategy.Analyze(stock, r.config)
		gridSignal := r.gridStrategy.Analyze(stock, r.config)

		result.WriteString(r.generateStockSection(stock, signal, gridSignal))
		result.WriteString("\n")

		switch signal.Action {
		case "BUY":
			buyCount++
		case "HOLD":
			holdCount++
		case "SELL":
			sellCount++
		case "WATCH":
			watchCount++
		}
	}

	// 汇总统计
	result.WriteString(r.generateSummary(len(stocks), buyCount, holdCount, sellCount, watchCount))

	// 资金监控和操作推荐
	if pool.AvailableFund > 0 {
		result.WriteString("\n")
		result.WriteString(r.generateFundRecommendation(pool, stocks))
	}

	// 免责声明
	result.WriteString("\n")
	result.WriteString("⚠️ 免责声明: 本工具仅供参考，不构成投资建议。投资有风险，入市需谨慎。\n")

	return result.String(), nil
}

// generateHeader 生成报告头部
func (r *Reporter) generateHeader() string {
	return fmt.Sprintf(`══════════════════════════════════════════════════════════════════
                    股息三 v1.0.0
                    分析报告 %s
══════════════════════════════════════════════════════════════════
`, time.Now().Format("2006-01-02 15:04"))
}

// generateConfigSummary 生成配置摘要
func (r *Reporter) generateConfigSummary() string {
	return fmt.Sprintf(`配置参数:
  默认目标股息率: %.2f%%
  最少分红年数: %d年
  便宜价折扣: %.0f%%
  昂贵价溢价: %.0f%%
`, r.config.DefaultTargetYield, r.config.MinDividendYears,
		r.config.CheapDiscount*100, r.config.ExpensivePremium*100)
}

// generateStockSection 生成股票分析部分
func (r *Reporter) generateStockSection(stock *model.Stock, signal *model.Signal, gridSignal *model.GridSignal) string {
	var result strings.Builder

	// 信号标识
	actionIcon := ""
	switch signal.Action {
	case "BUY":
		actionIcon = "买入 ★★★"
	case "HOLD":
		actionIcon = "持有 ★★"
	case "SELL":
		actionIcon = "卖出 ✗"
	case "WATCH":
		actionIcon = "观望 ★"
	}

	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))
	result.WriteString(fmt.Sprintf(" %s (%s)                              信号: %s\n", stock.Name, stock.Code, actionIcon))
	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))

	// 基本信息
	yieldStatus := "✓ 高于目标"
	if signal.CurrentYield < signal.TargetYield {
		yieldStatus = "✗ 低于目标"
	}

	result.WriteString(fmt.Sprintf("  当前价格:     ¥%.2f      涨跌幅: %.2f%%\n", stock.Price, stock.Change))
	result.WriteString(fmt.Sprintf("  每股股息:     ¥%.2f (2025年度)\n", stock.DividendPerShare))
	result.WriteString(fmt.Sprintf("  当前股息率:   %.2f%%      %s %.2f%%\n", signal.CurrentYield, yieldStatus, signal.CurrentYield-signal.TargetYield))
	result.WriteString(fmt.Sprintf("  连续分红:     %d年       %s\n", stock.DividendYears, r.checkDividendYears(stock.DividendYears)))
	result.WriteString("\n")

	// 估值区间
	result.WriteString(fmt.Sprintf("  ┌─ 估值区间 ─────────────────────────────────┐\n"))
	result.WriteString(fmt.Sprintf("  │ 便宜价:   ¥%.2f                          │\n", signal.CheapPrice))
	result.WriteString(fmt.Sprintf("  │ 合理价:   ¥%.2f                          │\n", signal.FairPrice))
	result.WriteString(fmt.Sprintf("  │ 昂贵价:   ¥%.2f                          │\n", signal.ExpensivePrice))

	if stock.Price <= signal.CheapPrice {
		result.WriteString(fmt.Sprintf("  │ 当前位置: 便宜价以下 (低估) ✓            │\n"))
	} else if stock.Price <= signal.FairPrice {
		result.WriteString(fmt.Sprintf("  │ 当前位置: 合理价以下 (合理)              │\n"))
	} else if stock.Price <= signal.ExpensivePrice {
		result.WriteString(fmt.Sprintf("  │ 当前位置: 合理价与昂贵价之间 (偏高)      │\n"))
	} else {
		result.WriteString(fmt.Sprintf("  │ 当前位置: 昂贵价以上 (高估) ✗            │\n"))
	}
	result.WriteString(fmt.Sprintf("  └────────────────────────────────────────────┘\n"))
	result.WriteString("\n")

	// 评分
	result.WriteString(fmt.Sprintf("  综合评分: %d/100\n", signal.Score))
	result.WriteString("\n")

	// 网格状态
	if gridSignal != nil {
		result.WriteString(r.gridStrategy.GetGridStatus(stock, r.config))
	}

	// 操作理由
	result.WriteString(fmt.Sprintf("  %s\n", signal.Reason))

	return result.String()
}

// checkDividendYears 检查分红年数
func (r *Reporter) checkDividendYears(years int) string {
	if years >= r.config.MinDividendYears {
		return "✓ 满足要求"
	}
	return "✗ 不满足要求"
}

// generateSummary 生成汇总统计
func (r *Reporter) generateSummary(total, buy, hold, sell, watch int) string {
	return fmt.Sprintf(`══════════════════════════════════════════════════════════════════
                           汇总统计
══════════════════════════════════════════════════════════════════
  监控股票:     %d只
  买入信号:     %d只
  持有信号:     %d只
  观望信号:     %d只
  卖出信号:     %d只
══════════════════════════════════════════════════════════════════
`, total, buy, hold, watch, sell)
}

// generateFundRecommendation 生成资金推荐
func (r *Reporter) generateFundRecommendation(pool *model.InvestPool, stocks map[string]*model.Stock) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))
	result.WriteString(fmt.Sprintf("                    资金监控与操作推荐\n"))
	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n\n"))

	result.WriteString(r.monitor.GetPoolSummary(pool))

	recommends := r.monitor.GenerateRecommendations(pool, stocks)

	result.WriteString(fmt.Sprintf("  ── 推荐操作 ──────────────────────────────────────────────────\n\n"))

	for i, rec := range recommends {
		if i >= 5 {
			break
		}

		priorityIcon := "○"
		if rec.Priority <= 2 {
			priorityIcon = "★"
		} else if rec.Priority <= 3 {
			priorityIcon = "◆"
		}

		switch rec.Action {
		case "BUY":
			result.WriteString(fmt.Sprintf("  %s 优先级%d: 买入 %s (%s)\n", priorityIcon, rec.Priority, rec.Stock.Name, rec.Stock.Code))
			result.WriteString(fmt.Sprintf("    建议投入: ¥%.0f\n", rec.Amount))
		case "SELL":
			result.WriteString(fmt.Sprintf("  %s 优先级%d: 卖出 %s (%s)\n", priorityIcon, rec.Priority, rec.Stock.Name, rec.Stock.Code))
			result.WriteString(fmt.Sprintf("    建议卖出: %.0f股\n", rec.Amount))
		case "HOLD":
			result.WriteString(fmt.Sprintf("  %s 优先级%d: 持有 %s (%s)\n", priorityIcon, rec.Priority, rec.Stock.Name, rec.Stock.Code))
		}

		if rec.Constraint != "" {
			result.WriteString(fmt.Sprintf("    ⚠️ %s\n", rec.Constraint))
		}
		result.WriteString(fmt.Sprintf("    理由: %s\n\n", rec.Reason))
	}

	// 资金分配建议
	totalBuyAmount := 0.0
	for _, rec := range recommends {
		if rec.Action == "BUY" {
			totalBuyAmount += rec.Amount
		}
	}

	result.WriteString(fmt.Sprintf("  ── 资金分配建议 ──────────────────────────────────────────────\n"))
	result.WriteString(fmt.Sprintf("  建议投入总额: ¥%.0f\n", totalBuyAmount))
	result.WriteString(fmt.Sprintf("  剩余可用资金: ¥%.0f\n", pool.AvailableFund-totalBuyAmount))
	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))

	return result.String()
}
