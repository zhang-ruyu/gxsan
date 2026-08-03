package cli

import (
	"fmt"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/model"
	"github.com/user/gxsan/internal/plan"
)

// yocOf 计算某持仓的成本股息率（无持仓返回0）
func yocOf(cfg *model.Config, code string, s *model.Stock) float64 {
	for _, h := range cfg.Portfolio {
		if h.Code == code && h.AvgCost > 0 {
			return data.CalculateDividendYield(h.AvgCost, s.DividendPerShare)
		}
	}
	return 0
}

// handleCompare 两只股票多维对比
func (a *App) handleCompare(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: gxsan compare <代码A> <代码B>")
		return
	}
	codeA, codeB := args[0], args[1]

	stocks := a.fetcher.EnrichStocks(a.cache, []string{codeA, codeB}, false)
	sa, okA := stocks[codeA]
	sb, okB := stocks[codeB]
	if !okA || !okB {
		fmt.Println("获取股票失败，请检查代码或网络")
		return
	}

	r := plan.CompareStocks(sa, sb, yocOf(a.configMgr.Config, codeA, sa), yocOf(a.configMgr.Config, codeB, sb))

	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Printf("                    个股对比: %s vs %s\n", r.A.Name, r.B.Name)
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Printf("  %-14s %-10s   %-10s\n", "维度", r.A.Name, r.B.Name)
	fmt.Printf("  %-14s %-10.2f%%   %-10.2f%%\n", "当前股息率", r.A.CurrentYield, r.B.CurrentYield)
	fmt.Printf("  %-14s %-10.2f    %-10.2f\n", "PE", r.A.PE, r.B.PE)
	fmt.Printf("  %-14s %-10.2f    %-10.2f\n", "PB", r.A.PB, r.B.PB)
	fmt.Printf("  %-14s %-10d年    %-10d年\n", "连续分红", r.A.DividendYears, r.B.DividendYears)
	if r.A.YieldOnCost > 0 || r.B.YieldOnCost > 0 {
		fmt.Printf("  %-14s %-10.2f%%   %-10.2f%%\n", "成本股息率", r.A.YieldOnCost, r.B.YieldOnCost)
	}
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Printf("  结论: %s 更优\n", r.Better)
	fmt.Printf("  依据: %s\n", r.Reason)
}
