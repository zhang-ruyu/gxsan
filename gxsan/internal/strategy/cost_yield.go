package strategy

import "github.com/user/gxsan/internal/data"

// CalculateYoC 成本股息率（Yield on Cost）= 每股分红 / 持仓成本 × 100
// 对应技能核心指标：通过低吸+网格降本持续压低持仓成本后，
// 成本股息率会远超静态股息率，是「攒股收息」质变的关键。
func CalculateYoC(avgCost, dividendPerShare float64) float64 {
	return data.CalculateDividendYield(avgCost, dividendPerShare)
}
