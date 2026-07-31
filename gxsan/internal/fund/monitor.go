package fund

import (
	"fmt"
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
