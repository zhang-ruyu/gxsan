package strategy

import (
	"fmt"
	"time"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/model"
)

// DividendStrategy 股息率策略
type DividendStrategy struct{}

// NewDividendStrategy 创建股息率策略
func NewDividendStrategy() *DividendStrategy {
	return &DividendStrategy{}
}

// Analyze 分析股票
func (s *DividendStrategy) Analyze(stock *model.Stock, config *model.Config) *model.Signal {
	// 获取目标股息率
	targetYield := config.DefaultTargetYield
	if sc := config.FindStock(stock.Code); sc != nil {
		targetYield = sc.TargetYield
	}

	// 计算股息率
	stock.DividendYield = data.CalculateDividendYield(stock.Price, stock.DividendPerShare)

	// 计算估值区间
	fairPrice := data.CalculateFairPrice(stock.DividendPerShare, targetYield)
	cheapPrice := fairPrice * config.CheapDiscount
	expensivePrice := fairPrice * config.ExpensivePremium

	// 计算评分
	score := s.calculateScore(stock, config, targetYield, fairPrice)

	// 生成信号
	signal := &model.Signal{
		Code:           stock.Code,
		Name:           stock.Name,
		Price:          stock.Price,
		CurrentYield:   stock.DividendYield,
		TargetYield:    targetYield,
		FairPrice:      fairPrice,
		CheapPrice:     cheapPrice,
		ExpensivePrice: expensivePrice,
		Score:          score,
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	}

	// 判断操作
	signal.Action, signal.Reason = s.determineAction(stock, config, signal)

	return signal
}

// calculateScore 计算评分
func (s *DividendStrategy) calculateScore(stock *model.Stock, config *model.Config, targetYield, fairPrice float64) int {
	score := 0

	// 股息率评分
	if stock.DividendYield >= targetYield {
		score += 30
		if stock.DividendYield >= targetYield*1.5 {
			score += 10
		}
	}

	// 分红年数评分
	if stock.DividendYears >= config.MinDividendYears {
		score += 20
		if stock.DividendYears >= 5 {
			score += 10
		}
	}

	// 价格位置评分
	if stock.Price <= fairPrice*config.CheapDiscount {
		score += 20
	} else if stock.Price <= fairPrice {
		score += 10
	}

	return score
}

// determineAction 判断操作
func (s *DividendStrategy) determineAction(stock *model.Stock, config *model.Config, signal *model.Signal) (string, string) {
	// 检查是否满足持仓要求
	if stock.DividendYears < config.MinDividendYears {
		return "WATCH", fmt.Sprintf("连续分红不足%d年，建议观察", config.MinDividendYears)
	}

	// 根据评分判断
	switch {
	case signal.Score >= 70:
		return "BUY", fmt.Sprintf("股息率%.2f%%高于目标%.2f%%，当前价格%.2f低于合理价%.2f，建议买入",
			signal.CurrentYield, signal.TargetYield, signal.Price, signal.FairPrice)
	case signal.Score >= 50:
		return "HOLD", fmt.Sprintf("股息率%.2f%%接近目标%.2f%%，建议持有",
			signal.CurrentYield, signal.TargetYield)
	case signal.Score >= 30:
		return "WATCH", fmt.Sprintf("股息率%.2f%%低于目标%.2f%%，建议观望",
			signal.CurrentYield, signal.TargetYield)
	default:
		return "SELL", fmt.Sprintf("股息率%.2f%%远低于目标%.2f%%，当前价格%.2f高于合理价%.2f，建议卖出",
			signal.CurrentYield, signal.TargetYield, signal.Price, signal.FairPrice)
	}
}
