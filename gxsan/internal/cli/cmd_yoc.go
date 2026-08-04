package cli

import (
	"fmt"

	"github.com/user/gxsan/internal/data"
)

// handleYoC 持仓成本股息率(YoC)
func (a *App) handleYoC(args []string) {
	if len(a.configMgr.Config.Portfolio) == 0 {
		fmt.Println("持仓为空，先用 `gxsan portfolio add` 添加持仓")
		return
	}

	// 并发抓取行情
	codes := make([]string, 0, len(a.configMgr.Config.Portfolio))
	for _, h := range a.configMgr.Config.Portfolio {
		codes = append(codes, h.Code)
	}
	stocks := a.fetcher.EnrichStocks(a.cache, codes, false)

	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println("                    持仓成本股息率(YoC)")
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Printf("  %-8s %-10s %-10s %-10s %-10s\n", "代码", "成本", "现价", "静态股息率", "YoC")
	fmt.Println("  ──────────────────────────────────────────────────────────")
	for _, h := range a.configMgr.Config.Portfolio {
		price := 0.0
		dps := 0.0
		if s, ok := stocks[h.Code]; ok {
			price = s.Price
			dps = s.DividendPerShare
		}
		staticY := data.CalculateDividendYield(price, dps)
		yoc := 0.0
		if h.AvgCost > 0 {
			yoc = data.CalculateDividendYield(h.AvgCost, dps)
		}
		fmt.Printf("  %-8s %-10.2f %-10.2f %-10.2f%% %-10.2f%%\n",
			h.Code, h.AvgCost, price, staticY, yoc)
	}
}
