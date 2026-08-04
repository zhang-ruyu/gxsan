package fund

import (
	"fmt"
	"math"
	"sort"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/model"
	"github.com/user/gxsan/internal/strategy"
)

// Monitor 资金监控器
type Monitor struct {
	config *model.Config
	cache  *data.Cache
}

// NewMonitor 创建资金监控器
func NewMonitor(config *model.Config, cache *data.Cache) *Monitor {
	return &Monitor{
		config: config,
		cache:  cache,
	}
}

// BuildPool 构建投资池
func (m *Monitor) BuildPool(stocks map[string]*model.Stock) *model.InvestPool {
	holdings := make(map[string]model.Holding)

	// 从配置中加载持仓
	for _, h := range m.config.Portfolio {
		// 更新持仓的当前价格
		if stock, ok := stocks[h.Code]; ok {
			h.CurrentPrice = stock.Price
			h.MarketValue = float64(h.Shares) * stock.Price
			h.Dividend = stock.DividendPerShare * float64(h.Shares)
			h.YieldOnCost = data.CalculateDividendYield(h.AvgCost, stock.DividendPerShare)
		}
		holdings[h.Code] = h
	}

	pool := &model.InvestPool{
		AvailableFund:  m.config.Fund.AvailableFund,
		Holdings:       holdings,
		MaxPositionPct: m.config.Fund.MaxPositionPct,
	}

	pool.CalculatePool()
	return pool
}

// GenerateRecommendations 生成操作推荐
func (m *Monitor) GenerateRecommendations(pool *model.InvestPool, stocks map[string]*model.Stock) []model.Recommend {
	var recommends []model.Recommend

	gridStrategy := strategy.NewGridStrategy()

	for _, stockConfig := range m.config.Watchlist {
		stock, ok := stocks[stockConfig.Code]
		if !ok {
			continue
		}

		// 获取网格信号
		gridSignal := gridStrategy.Analyze(stock, m.config)
		if gridSignal == nil {
			continue
		}

		// 获取股息策略信号
		divStrategy := strategy.NewDividendStrategy()
		signal := divStrategy.Analyze(stock, m.config)

		recommend := model.Recommend{
			Stock:    *stock,
			Priority: 5,
		}

		switch gridSignal.Action {
		case "BUY":
			recommend.Action = "BUY"
			recommend.Amount = gridSignal.BuyAmount

			// 检查持仓约束
			currentHolding := pool.Holdings[stock.Code]
			currentValue := currentHolding.MarketValue
			maxAllowed := pool.MaxPosition - currentValue

			if recommend.Amount > maxAllowed {
				recommend.Amount = maxAllowed
				recommend.Constraint = fmt.Sprintf("已达单只股票持仓上限%.0f%%", m.config.Fund.MaxPositionPct)
			}

			if recommend.Amount > pool.AvailableFund {
				recommend.Amount = pool.AvailableFund
				recommend.Constraint = "可用资金不足"
			}

			if recommend.Amount <= 0 {
				recommend.Constraint = "已达持仓上限或资金不足"
				recommend.Priority = 5
			} else {
				recommend.Priority = 6 - gridSignal.BuyLevel
			}

			recommend.Reason = fmt.Sprintf("当前股息率%.2f%%（目标%.2f%%），网格建议投入%.0f元",
				signal.CurrentYield, signal.TargetYield, recommend.Amount)

		case "SELL":
			recommend.Action = "SELL"
			currentHolding := pool.Holdings[stock.Code]
			if currentHolding.Shares > 0 {
				sellShares := int(float64(currentHolding.Shares) * gridSignal.SellPercent / 100)
				if sellShares == 0 {
					sellShares = 1
				}
				recommend.Amount = float64(sellShares)
				recommend.Priority = gridSignal.SellLevel
				recommend.Reason = fmt.Sprintf("当前股息率%.2f%%触发卖出，建议卖出%d股（%.0f%%）",
					signal.CurrentYield, sellShares, gridSignal.SellPercent)
			} else {
				continue
			}

		case "HOLD":
			recommend.Action = "HOLD"
			recommend.Priority = 5
			recommend.Reason = fmt.Sprintf("当前股息率%.2f%%，建议持有", signal.CurrentYield)
		}

		recommends = append(recommends, recommend)
	}

	// 按优先级排序
	sort.Slice(recommends, func(i, j int) bool {
		return recommends[i].Priority < recommends[j].Priority
	})

	return recommends
}

// GenerateHoldingRecommendations 为「持仓组合」生成行动提示（建议引擎 MVP 核心，
// 复用现有 divStrategy + gridStrategy）。
// 与 GenerateRecommendations（仅跑监控列表）不同，本方法遍历 config.Portfolio，
// 对每只持仓给出 买/卖/等 行动 + 建议金额 + 触发理由。
func (m *Monitor) GenerateHoldingRecommendations(pool *model.InvestPool, stocks map[string]*model.Stock) []model.ActionAdvice {
	divStrategy := strategy.NewDividendStrategy()
	gridStrategy := strategy.NewGridStrategy()

	var advices []model.ActionAdvice
	for _, h := range m.config.Portfolio {
		stock, ok := stocks[h.Code]
		if !ok {
			continue
		}

		divSig := divStrategy.Analyze(stock, m.config)
		gridSig := gridStrategy.Analyze(stock, m.config) // 未配置网格时返回 nil

		holding := pool.Holdings[h.Code]
		currentValue := holding.MarketValue
		maxAllowed := pool.MaxPosition - currentValue
		if maxAllowed < 0 {
			maxAllowed = 0
		}

		advice := model.ActionAdvice{
			Code:             h.Code,
			Name:             h.Name,
			Shares:           h.Shares,
			AvgCost:          h.AvgCost,
			Price:            stock.Price,
			YieldOnCost:      holding.YieldOnCost,
			CurrentYield:     divSig.CurrentYield,
			TargetYield:      divSig.TargetYield,
			CheapPrice:       divSig.CheapPrice,
			FairPrice:        divSig.FairPrice,
			CostCorrectable:   h.TotalCost == 0, // 尚无真实投入总额记录 → 提示可一键修正
		}

		// 默认采用股息策略信号（价值维度：BUY/HOLD/WATCH/SELL）
		action := divSig.Action
		reason := divSig.Reason
		priority := 5
		switch action {
		case "BUY":
			priority = 2
		case "SELL":
			priority = 1
		default: // HOLD / WATCH
			priority = 5
		}

		// 若配置了网格，则以网格信号为准（含更具体的金额）
		if gridSig != nil {
			action = gridSig.Action
			reason = gridSig.Reason
			switch gridSig.Action {
			case "BUY":
				priority = 6 - gridSig.BuyLevel
			case "SELL":
				priority = gridSig.SellLevel
			default:
				priority = 5
			}
		}

		advice.Action = action
		advice.Reason = reason
		advice.Priority = priority

		switch action {
		case "BUY":
			amount := 0.0
			if gridSig != nil {
				amount = gridSig.BuyAmount
			} else {
				// 未设网格：按可用资金 / 持仓上限给出一个默认买入额
				amount = math.Min(m.config.Fund.AvailableFund, maxAllowed)
				advice.Constraint = "未配置网格，按可用资金/持仓上限建议"
			}
			// 资金与持仓上限约束
			if amount > m.config.Fund.AvailableFund {
				amount = m.config.Fund.AvailableFund
				advice.Constraint = "可用资金不足"
			}
			if amount > maxAllowed {
				amount = maxAllowed
				advice.Constraint = fmt.Sprintf("已达单只股票持仓上限%.0f%%", m.config.Fund.MaxPositionPct)
			}
			advice.SuggestedBuyAmount = amount
		case "SELL":
			pct := 0.0
			if gridSig != nil {
				pct = gridSig.SellPercent
			} else {
				pct = 20 // 无网格时温和减仓 20%
			}
			sellShares := int(float64(h.Shares) * pct / 100)
			if sellShares == 0 {
				sellShares = 1
			}
			if sellShares > h.Shares {
				sellShares = h.Shares
			}
			advice.SuggestedSellShares = sellShares
		}

		advices = append(advices, advice)
	}

	sort.Slice(advices, func(i, j int) bool {
		if advices[i].Priority != advices[j].Priority {
			return advices[i].Priority < advices[j].Priority
		}
		return advices[i].Code < advices[j].Code
	})

	return advices
}

// GetPoolSummary 获取投资池摘要
func (m *Monitor) GetPoolSummary(pool *model.InvestPool) string {
	result := fmt.Sprintf("  资金状况:\n")
	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")
	result += fmt.Sprintf("  可用资金:       ¥%.0f\n", pool.AvailableFund)

	totalMarketValue := 0.0
	for _, h := range pool.Holdings {
		totalMarketValue += h.MarketValue
	}

	result += fmt.Sprintf("  持仓市值:       ¥%.0f\n", totalMarketValue)
	result += fmt.Sprintf("  总资产:         ¥%.0f\n", pool.TotalAsset)
	result += fmt.Sprintf("  单只股票上限:   ¥%.0f (%.0f%%)\n", pool.MaxPosition, pool.MaxPositionPct)
	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")

	return result
}
