package cli

import (
	"fmt"
	"strconv"
)

// handlePortfolio 处理投资组合
func (a *App) handlePortfolio(args []string) {
	if len(args) < 1 {
		fmt.Println("用法:")
		fmt.Println("  gxsan portfolio add <代码> --shares <股数> --cost <成本价>")
		fmt.Println("  gxsan portfolio update <代码> --shares <股数>")
		fmt.Println("  gxsan portfolio list")
		fmt.Println("  gxsan portfolio remove <代码>")
		return
	}

	switch args[0] {
	case "add":
		a.handlePortfolioAdd(args[1:])
	case "update":
		a.handlePortfolioUpdate(args[1:])
	case "list":
		a.handlePortfolioList()
	case "remove":
		a.handlePortfolioRemove(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

// handlePortfolioAdd 处理添加持仓
func (a *App) handlePortfolioAdd(args []string) {
	if len(args) < 5 {
		fmt.Println("用法: gxsan portfolio add <代码> --shares <股数> --cost <成本价>")
		return
	}

	code := args[0]
	shares := 0
	cost := 0.0
	name := ""

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--shares":
			if i+1 < len(args) {
				shares, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "--cost":
			if i+1 < len(args) {
				cost, _ = strconv.ParseFloat(args[i+1], 64)
				i++
			}
		}
	}

	// 查找股票名称
	for _, stock := range a.configMgr.Config.Watchlist {
		if stock.Code == code {
			name = stock.Name
			break
		}
	}

	if name == "" {
		name = code
	}

	if err := a.configMgr.AddHolding(code, name, shares, cost); err != nil {
		fmt.Printf("添加失败: %v\n", err)
		return
	}

	fmt.Printf("已添加持仓: %s (%s), %d股, 成本价 ¥%.2f\n", name, code, shares, cost)
}

// handlePortfolioUpdate 处理更新持仓
func (a *App) handlePortfolioUpdate(args []string) {
	if len(args) < 3 {
		fmt.Println("用法: gxsan portfolio update <代码> --shares <股数>")
		return
	}

	code := args[0]
	shares := 0

	for i := 1; i < len(args); i++ {
		if args[i] == "--shares" && i+1 < len(args) {
			shares, _ = strconv.Atoi(args[i+1])
			i++
		}
	}

	if err := a.configMgr.UpdateHolding(code, shares); err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}

	fmt.Printf("已更新持仓: %s, %d股\n", code, shares)
}

// handlePortfolioList 处理持仓列表
func (a *App) handlePortfolioList() {
	if len(a.configMgr.Config.Portfolio) == 0 {
		fmt.Println("持仓列表为空")
		return
	}

	fmt.Println("持仓列表:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("%-8s %-10s %-8s %-10s\n", "代码", "名称", "股数", "成本价")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for _, holding := range a.configMgr.Config.Portfolio {
		fmt.Printf("%-8s %-10s %-8d ¥%.4f\n", holding.Code, holding.Name, holding.Shares, holding.AvgCost)
	}
}

// handlePortfolioRemove 处理删除持仓
func (a *App) handlePortfolioRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: gxsan portfolio remove <代码>")
		return
	}

	code := args[0]
	if err := a.configMgr.RemoveHolding(code); err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	fmt.Printf("已删除持仓: %s\n", code)
}
