// Package plan 攒股收息的「规划测算」领域层。
// 与 strategy（单只股票的买卖信号）不同，plan 聚焦长期目标推导：
// 定投复利模拟、养老现金流倒推、资产转换比价。
package plan

// CompoundYear 定投复利模拟的逐年结果
type CompoundYear struct {
	Year            int     `json:"year"`
	TotalAsset      float64 `json:"total_asset"`      // 年末总资产
	AnnualDividend  float64 `json:"annual_dividend"`  // 年末按股息率估算的年股息
	MonthlyDividend float64 `json:"monthly_dividend"` // 月均股息
}

// CompoundSchedule 定投复利模拟
// monthlyInvest: 每月定投金额；years: 年数；annualRate: 年化收益率(小数, 如0.05)
// 模型：每月资产按月复利增长 annualRate/12，并追加当月定投；分红100%复投已内含于复利。
// 年末年股息 = 总资产 × annualRate（即当年股息率）。
// 对应技能「定投复利模拟」：月定投X元、买入股息率R资产、分红100%复投。
func CompoundSchedule(monthlyInvest float64, years int, annualRate float64) []CompoundYear {
	rows := make([]CompoundYear, 0, years)
	asset := 0.0
	monthlyRate := annualRate / 12
	for y := 1; y <= years; y++ {
		for m := 0; m < 12; m++ {
			asset = asset*(1+monthlyRate) + monthlyInvest
		}
		annualDiv := asset * annualRate
		rows = append(rows, CompoundYear{
			Year:            y,
			TotalAsset:      asset,
			AnnualDividend:  annualDiv,
			MonthlyDividend: annualDiv / 12,
		})
	}
	return rows
}
