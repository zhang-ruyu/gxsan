package strategy

import (
	"fmt"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/model"
)

// GridStrategy 网格策略
type GridStrategy struct{}

// NewGridStrategy 创建网格策略
func NewGridStrategy() *GridStrategy {
	return &GridStrategy{}
}

// Analyze 分析网格信号（基于持仓/监控配置中的网格档位）
func (s *GridStrategy) Analyze(stock *model.Stock, config *model.Config) *model.GridSignal {
	// 获取股票配置
	stockConfig := config.FindStock(stock.Code)
	if stockConfig == nil {
		return nil
	}

	// 计算股息率
	yield := data.CalculateDividendYield(stock.Price, stock.DividendPerShare)

	return s.AnalyzeYield(yield, stockConfig.GridStrategy.BuyGrids, stockConfig.GridStrategy.SellGrids)
}

// AnalyzeYield 给定当前股息率与一组买入/卖出网格，计算触发档位与建议动作。
// 与 Analyze 同算法，但不依赖 config（持仓可直接用自身网格档位调用，无需伪造监控项）。
func (s *GridStrategy) AnalyzeYield(yield float64, buyGrids [5]model.GridLevel, sellGrids [5]model.GridLevel) *model.GridSignal {
	signal := &model.GridSignal{
		CurrentYield: yield,
	}

	// 分析买入网格
	signal.BuyLevel, signal.BuyAmount = s.analyzeBuyGrid(yield, buyGrids)

	// 分析卖出网格
	signal.SellLevel, signal.SellPercent = s.analyzeSellGrid(yield, sellGrids)

	// 判断操作
	signal.Action, signal.Reason = s.determineAction(signal)

	return signal
}

// analyzeBuyGrid 分析买入网格
func (s *GridStrategy) analyzeBuyGrid(yield float64, grids [5]model.GridLevel) (int, float64) {
	level := 0
	amount := 0.0

	// 从高到低检查，找到当前满足的最高档位
	for i := 4; i >= 0; i-- {
		if yield >= grids[i].Yield {
			level = i + 1
			amount = grids[i].Amount
			break
		}
	}

	return level, amount
}

// analyzeSellGrid 分析卖出网格
func (s *GridStrategy) analyzeSellGrid(yield float64, grids [5]model.GridLevel) (int, float64) {
	level := 0
	percent := 0.0

	// 从低到高检查，找到当前满足的最低档位
	for i := 0; i < 5; i++ {
		if yield <= grids[i].Yield {
			level = i + 1
			percent = grids[i].Amount
			break
		}
	}

	return level, percent
}

// determineAction 判断操作
func (s *GridStrategy) determineAction(signal *model.GridSignal) (string, string) {
	if signal.BuyLevel > 0 && signal.SellLevel > 0 {
		// 同时满足买入和卖出条件，优先卖出
		return "SELL", fmt.Sprintf("当前股息率%.2f%%触发卖出网格第%d档，建议卖出%.0f%%",
			signal.CurrentYield, signal.SellLevel, signal.SellPercent)
	}

	if signal.BuyLevel > 0 {
		return "BUY", fmt.Sprintf("当前股息率%.2f%%触发买入网格第%d档，建议投入%.0f元",
			signal.CurrentYield, signal.BuyLevel, signal.BuyAmount)
	}

	if signal.SellLevel > 0 {
		return "SELL", fmt.Sprintf("当前股息率%.2f%%触发卖出网格第%d档，建议卖出%.0f%%",
			signal.CurrentYield, signal.SellLevel, signal.SellPercent)
	}

	return "HOLD", fmt.Sprintf("当前股息率%.2f%%未触发任何网格，建议持有", signal.CurrentYield)
}

// GetGridStatus 获取网格状态
func (s *GridStrategy) GetGridStatus(stock *model.Stock, config *model.Config) string {
	stockConfig := config.FindStock(stock.Code)
	if stockConfig == nil {
		return ""
	}

	yield := data.CalculateDividendYield(stock.Price, stock.DividendPerShare)

	result := fmt.Sprintf("  买入网格:\n")
	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")
	result += fmt.Sprintf("  档位   股息率阈值   建议投入    当前状态\n")
	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")

	for i, grid := range stockConfig.GridStrategy.BuyGrids {
		status := "○ 待达到"
		if yield >= grid.Yield {
			if i+1 == s.getCurrentBuyLevel(yield, stockConfig.GridStrategy.BuyGrids) {
				status = "★ 当前位置"
			} else {
				status = "✓ 已超过"
			}
		}
		result += fmt.Sprintf("  %d      >= %.1f%%     ¥%.0f     %s\n", i+1, grid.Yield, grid.Amount, status)
	}

	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n\n")

	result += fmt.Sprintf("  卖出网格:\n")
	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")
	result += fmt.Sprintf("  档位   股息率阈值   卖出比例    当前状态\n")
	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")

	for i, grid := range stockConfig.GridStrategy.SellGrids {
		status := "○ 未触发"
		if yield <= grid.Yield {
			status = "★ 当前位置"
		}
		result += fmt.Sprintf("  %d      <= %.1f%%     %.0f%%       %s\n", i+1, grid.Yield, grid.Amount, status)
	}

	result += fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n")

	return result
}

// getCurrentBuyLevel 获取当前买入档位
func (s *GridStrategy) getCurrentBuyLevel(yield float64, grids [5]model.GridLevel) int {
	for i := 4; i >= 0; i-- {
		if yield >= grids[i].Yield {
			return i + 1
		}
	}
	return 0
}
