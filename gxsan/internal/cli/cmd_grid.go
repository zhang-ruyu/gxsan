package cli

import (
	"fmt"
	"strconv"
)

// handleGrid 处理网格
func (a *App) handleGrid(args []string) {
	if len(args) < 1 {
		fmt.Println("用法:")
		fmt.Println("  gxsan grid set <代码> --buy <档位> <股息率> <金额>")
		fmt.Println("  gxsan grid set <代码> --sell <档位> <股息率> <比例>")
		fmt.Println("  gxsan grid show <代码>")
		return
	}

	switch args[0] {
	case "set":
		a.handleGridSet(args[1:])
	case "show":
		a.handleGridShow(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

// handleGridSet 处理网格设置
func (a *App) handleGridSet(args []string) {
	if len(args) < 5 {
		fmt.Println("用法: gxsan grid set <代码> --buy/--sell <档位> <股息率> <金额/比例>")
		return
	}

	code := args[0]
	isBuy := args[1] == "--buy"
	level, _ := strconv.Atoi(args[2])
	yield, _ := strconv.ParseFloat(args[3], 64)
	amount, _ := strconv.ParseFloat(args[4], 64)

	if err := a.configMgr.SetGrid(code, isBuy, level, yield, amount); err != nil {
		fmt.Printf("设置失败: %v\n", err)
		return
	}

	gridType := "买入"
	if !isBuy {
		gridType = "卖出"
	}
	amountType := "金额"
	if !isBuy {
		amountType = "比例"
	}
	fmt.Printf("已设置 %s %s 网格第%d档: 股息率 %.2f%%, %s %.0f\n", code, gridType, level, yield, amountType, amount)
}

// handleGridShow 处理网格显示
func (a *App) handleGridShow(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: gxsan grid show <代码>")
		return
	}

	code := args[0]
	stockConfig := a.configMgr.GetStockConfig(code)
	if stockConfig == nil {
		fmt.Printf("股票 %s 不存在\n", code)
		return
	}

	fmt.Printf("网格策略 - %s (%s)\n", stockConfig.Name, stockConfig.Code)
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("买入网格:")
	fmt.Println("  档位   股息率阈值   建议投入")
	fmt.Println("  ──────────────────────────────────────────────────────────")
	for i, grid := range stockConfig.GridStrategy.BuyGrids {
		fmt.Printf("  %d      >= %.1f%%     ¥%.0f\n", i+1, grid.Yield, grid.Amount)
	}
	fmt.Println()
	fmt.Println("卖出网格:")
	fmt.Println("  档位   股息率阈值   卖出比例")
	fmt.Println("  ──────────────────────────────────────────────────────────")
	for i, grid := range stockConfig.GridStrategy.SellGrids {
		fmt.Printf("  %d      <= %.1f%%     %.0f%%\n", i+1, grid.Yield, grid.Amount)
	}
}
