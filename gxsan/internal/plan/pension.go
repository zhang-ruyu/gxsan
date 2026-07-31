package plan

// PensionRow 目标倒推的所需本金（不同成本股息率下）
type PensionRow struct {
	CostYield         float64 `json:"cost_yield"`          // 成本股息率
	RequiredPrincipal float64 `json:"required_principal"`  // 所需本金
}

// RequiredPrincipal 目标倒推公式：所需本金 = 目标年分红 / 成本股息率
// 对应技能「养老现金流测算 · 目标倒推公式」。
func RequiredPrincipal(targetAnnualDividend, costYield float64) float64 {
	if costYield <= 0 {
		return 0
	}
	return targetAnnualDividend / costYield
}

// PensionPlan 根据目标月分红，倒推不同成本股息率(默认5/6/7%)下的所需本金
func PensionPlan(targetMonthlyDividend float64, costYields []float64) []PensionRow {
	annual := targetMonthlyDividend * 12
	rows := make([]PensionRow, 0, len(costYields))
	for _, cy := range costYields {
		rows = append(rows, PensionRow{
			CostYield:         cy,
			RequiredPrincipal: RequiredPrincipal(annual, cy),
		})
	}
	return rows
}

// RetirementNote 退休提款阶段说明（按生命周期阶段给出策略提示）
func RetirementNote(stage int) string {
	switch stage {
	case 1:
		return "启动期：坚持定投、养成复投习惯，不急于看效果。"
	case 2:
		return "滚雪球期：坚决不提现，分红100%复投，下跌加仓。"
	case 3:
		return "自由期：部分消费+部分复投，克制消费冲动。"
	case 4:
		return "收获期：提取改善生活，其余复投，考虑传承；提取率建议≤5%。"
	default:
		return "滚雪球期：坚决不提现，分红100%复投。"
	}
}
