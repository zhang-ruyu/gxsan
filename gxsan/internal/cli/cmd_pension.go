package cli

import (
	"fmt"
	"strconv"

	"github.com/user/gxsan/internal/model"
	"github.com/user/gxsan/internal/plan"
)

// handlePension 养老现金流测算：目标倒推 + 定投复利模拟 + 退休提款
func (a *App) handlePension(args []string) {
	monthly := 5000.0 // 目标月分红
	invest := 3000.0  // 月定投
	years := 20
	rate := 0.05 // 年化 5%

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--monthly":
			if i+1 < len(args) {
				if v, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					monthly = v
				}
				i++
			}
		case "--invest":
			if i+1 < len(args) {
				if v, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					invest = v
				}
				i++
			}
		case "--years":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					years = v
				}
				i++
			}
		case "--rate":
			if i+1 < len(args) {
				if v, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					rate = v / 100
				}
				i++
			}
		}
	}

	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println("                    养老现金流测算")
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Printf("生命周期阶段: %s\n", model.LifecycleStageName(a.configMgr.Config.LifecycleStage))
	fmt.Println()

	fmt.Println("【目标倒推】所需本金 = 目标年分红 ÷ 成本股息率")
	fmt.Printf("  目标月分红: ¥%.0f  →  目标年分红: ¥%.0f\n", monthly, monthly*12)
	for _, r := range plan.PensionPlan(monthly, []float64{0.05, 0.06, 0.07}) {
		fmt.Printf("  成本股息率 %.0f%%  →  所需本金 ¥%.0f\n", r.CostYield*100, r.RequiredPrincipal)
	}
	fmt.Println()

	fmt.Printf("【定投复利模拟】月定投 ¥%.0f × %d年 × 年化 %.0f%%（分红100%%复投）\n", invest, years, rate*100)
	for _, s := range plan.CompoundSchedule(invest, years, rate) {
		if s.Year == 1 || s.Year%5 == 0 {
			fmt.Printf("  第%02d年: 资产 ¥%.0f | 年股息 ¥%.0f | 月均 ¥%.0f\n", s.Year, s.TotalAsset, s.AnnualDividend, s.MonthlyDividend)
		}
	}
	fmt.Println()

	fmt.Printf("【退休提款策略】%s\n", plan.RetirementNote(a.configMgr.Config.LifecycleStage))
}
