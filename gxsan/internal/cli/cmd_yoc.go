package cli

import (
	"fmt"

	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/model"
)

// handleYoC 持仓成本股息率(YoC)与账户分组
func (a *App) handleYoC(args []string) {
	if len(a.configMgr.Config.Portfolio) == 0 {
		fmt.Println("持仓为空，先用 `gxsan portfolio add` 添加持仓")
		return
	}

	// 抓取行情
	stocks := make(map[string]*model.Stock)
	for _, h := range a.configMgr.Config.Portfolio {
		if s, err := a.fetcher.EnrichStock(a.cache, h.Code, false); err == nil {
			stocks[h.Code] = s
		}
	}

	// 按账户分组
	groups := make(map[string][]model.Holding)
	for _, h := range a.configMgr.Config.Portfolio {
		key := h.Account
		if key == "" {
			key = "未分配"
		}
		groups[key] = append(groups[key], h)
	}

	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println("              持仓成本股息率(YoC) · 按账户分组")
	fmt.Println("══════════════════════════════════════════════════════════════════")
	for acc, hs := range groups {
		fmt.Printf("\n【%s】\n", acc)
		fmt.Printf("  %-8s %-10s %-10s %-10s %-10s\n", "代码", "成本", "现价", "静态股息率", "YoC")
		fmt.Println("  ──────────────────────────────────────────────────────────")
		for _, h := range hs {
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
}
