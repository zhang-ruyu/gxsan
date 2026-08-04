package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/model"
)

// GenerateStockDetail 生成单只股票详情（JSON 友好的 model.StockDetail）。
// forceRefresh 为 true 时跳过缓存，重新抓数（供 GUI 刷新按钮使用）。
func (r *Reporter) GenerateStockDetail(code string, forceRefresh bool) (*model.StockDetail, error) {
	stock, err := r.fetcher.EnrichStock(r.cache, code, forceRefresh)
	if err != nil {
		return nil, fmt.Errorf("获取股票数据失败: %w", err)
	}

	signal := r.divStrategy.Analyze(stock, r.config)

	// 若该股在持仓中，补充成本股息率(YoC)
	var yoc float64
	for _, h := range r.config.Portfolio {
		if h.Code == code && h.AvgCost > 0 {
			yoc = data.CalculateDividendYield(h.AvgCost, stock.DividendPerShare)
			break
		}
	}

	// 历史分红（优先读缓存，未命中再补抓一次）
	var dividends []model.DividendRecord
	if cd, cerr := r.cache.Get(code); cerr == nil {
		dividends = cd.Dividends
	}
	if len(dividends) == 0 {
		dividends, _ = r.fetcher.GetDividendHistory(code)
	}
	history := buildHistory(dividends, stock.Price)

	// 近3年平均静态股息率 = (近3年每股分红之和/3)/现价
	avgYield := signal.CurrentYield
	if len(history) > 0 {
		sum, n := 0.0, 0
		for i := len(history) - 1; i >= 0 && n < 3; i-- {
			sum += history[i].DividendPerShare
			n++
		}
		if n > 0 && stock.Price > 0 {
			avgYield = (sum / float64(n)) / stock.Price * 100
		}
	}

	var gridInfo *model.GridStrategyInfo
	if sc := r.config.FindStock(code); sc != nil {
		gridInfo = &model.GridStrategyInfo{
			BuyGrids:  sc.GridStrategy.BuyGrids[:],
			SellGrids: sc.GridStrategy.SellGrids[:],
		}
	}

	return &model.StockDetail{
		StockCode:           stock.Code,
		StockName:           stock.Name,
		CurrentPrice:        stock.Price,
		Change:              stock.Change,
		PE:                  stock.PE,
		PB:                  stock.PB,
		LatestDividendYield: signal.CurrentYield,
		AvgDividendYield:    avgYield,
		TargetYield:         signal.TargetYield,
		DividendPerShare:    stock.DividendPerShare,
		DividendYears:       stock.DividendYears,
		YieldOnCost:         yoc,
		Signal:              signal.Action,
		Valuation:           "",
		Score:               signal.Score,
		Reason:              signal.Reason,
		CheapPrice:          signal.CheapPrice,
		FairPrice:           signal.FairPrice,
		ExpensivePrice:      signal.ExpensivePrice,
		CheapDiscount:       r.config.CheapDiscount,
		ExpensivePremium:    r.config.ExpensivePremium,
		GridStrategy:        gridInfo,
		History:             history,
	}, nil
}

// buildHistory 将分红记录转为按年历史项（最近在前）
func buildHistory(dividends []model.DividendRecord, price float64) []model.DividendHistoryItem {
	var items []model.DividendHistoryItem
	for _, d := range dividends {
		var y int
		fmt.Sscanf(d.Date, "%d-", &y)
		item := model.DividendHistoryItem{Year: y, DividendPerShare: d.Amount}
		if price > 0 {
			item.Yield = d.Amount / price * 100
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Year > items[j].Year })
	return items
}

// FormatStockDetailText 将详情渲染为 CLI 可读文本
func FormatStockDetailText(d *model.StockDetail) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))
	b.WriteString(fmt.Sprintf("                    %s (%s) 详细分析\n", d.StockName, d.StockCode))
	b.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n\n"))

	b.WriteString("基本信息\n")
	b.WriteString("────────────────────────────────────────────────────────────────\n")
	b.WriteString(fmt.Sprintf("  股票代码:     %s\n", d.StockCode))
	b.WriteString(fmt.Sprintf("  股票名称:     %s\n", d.StockName))
	b.WriteString(fmt.Sprintf("  当前价格:     ¥%.2f\n", d.CurrentPrice))
	b.WriteString(fmt.Sprintf("  今日涨跌:     %.2f%%\n", d.Change))
	b.WriteString(fmt.Sprintf("  市盈率(PE):   %.2f\n", d.PE))
	b.WriteString(fmt.Sprintf("  市净率(PB):   %.2f\n", d.PB))
	b.WriteString("\n")

	b.WriteString("股息数据\n")
	b.WriteString("────────────────────────────────────────────────────────────────\n")
	b.WriteString(fmt.Sprintf("  每股股息:     ¥%.2f\n", d.DividendPerShare))
	b.WriteString(fmt.Sprintf("  当前股息率:   %.2f%%\n", d.LatestDividendYield))
	b.WriteString(fmt.Sprintf("  平均股息率:   %.2f%%\n", d.AvgDividendYield))
	b.WriteString(fmt.Sprintf("  目标股息率:   %.2f%%\n", d.TargetYield))
	b.WriteString(fmt.Sprintf("  连续分红:     %d年\n", d.DividendYears))
	if d.YieldOnCost > 0 {
		b.WriteString(fmt.Sprintf("  成本股息率:   %.2f%% (YoC)\n", d.YieldOnCost))
	}
	b.WriteString("\n")

	b.WriteString("估值分析\n")
	b.WriteString("────────────────────────────────────────────────────────────────\n")
	b.WriteString(fmt.Sprintf("  便宜价:       ¥%.2f\n", d.CheapPrice))
	b.WriteString(fmt.Sprintf("  合理价:       ¥%.2f\n", d.FairPrice))
	b.WriteString(fmt.Sprintf("  昂贵价:       ¥%.2f\n", d.ExpensivePrice))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("综合评分: %d/100\n", d.Score))
	b.WriteString(fmt.Sprintf("操作建议: %s\n", d.Signal))
	b.WriteString(fmt.Sprintf("理由: %s\n", d.Reason))
	b.WriteString("\n")

	if d.GridStrategy != nil {
		b.WriteString("网格策略\n")
		b.WriteString("────────────────────────────────────────────────────────────────\n")
		b.WriteString("  买入网格:\n")
		for i, g := range d.GridStrategy.BuyGrids {
			b.WriteString(fmt.Sprintf("    档位%d: 股息率≥%.2f%% → 投入¥%.0f\n", i+1, g.Yield, g.Amount))
		}
		b.WriteString("  卖出网格:\n")
		for i, g := range d.GridStrategy.SellGrids {
			b.WriteString(fmt.Sprintf("    档位%d: 股息率≤%.2f%% → 卖出%.0f%%\n", i+1, g.Yield, g.Amount))
		}
		b.WriteString("\n")
	}

	if len(d.History) > 0 {
		b.WriteString("历史分红\n")
		b.WriteString("────────────────────────────────────────────────────────────────\n")
		for _, h := range d.History {
			b.WriteString(fmt.Sprintf("  %d年: 每股¥%.2f  静态股息率%.2f%%\n", h.Year, h.DividendPerShare, h.Yield))
		}
	}

	return b.String()
}
