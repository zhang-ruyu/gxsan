package cli

import (
	"fmt"
	"strconv"
)

// handleConfig 处理配置
func (a *App) handleConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("用法:")
		fmt.Println("  gxsan config show")
		fmt.Println("  gxsan config set <键> <值>")
		return
	}

	switch args[0] {
	case "show":
		a.handleConfigShow()
	case "set":
		a.handleConfigSet(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

// handleConfigShow 处理配置显示
func (a *App) handleConfigShow() {
	fmt.Println("当前配置:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("默认目标股息率:   %.2f%%\n", a.configMgr.Config.DefaultTargetYield)
	fmt.Printf("最少分红年数:     %d年\n", a.configMgr.Config.MinDividendYears)
	fmt.Printf("便宜价折扣:       %.0f%%\n", a.configMgr.Config.CheapDiscount*100)
	fmt.Printf("昂贵价溢价:       %.0f%%\n", a.configMgr.Config.ExpensivePremium*100)
	fmt.Printf("可用资金:         ¥%.0f\n", a.configMgr.Config.Fund.AvailableFund)
	fmt.Printf("单只股票上限:     %.0f%%\n", a.configMgr.Config.Fund.MaxPositionPct)
	fmt.Println("─────────────────────────────────────────────────────────────")
}

// handleConfigSet 处理配置设置
func (a *App) handleConfigSet(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: gxsan config set <键> <值>")
		fmt.Println("可用键:")
		fmt.Println("  default-target-yield  默认目标股息率")
		fmt.Println("  min-dividend-years    最少分红年数")
		fmt.Println("  cheap-discount        便宜价折扣")
		fmt.Println("  expensive-premium     昂贵价溢价")
		fmt.Println("  available-fund        可用资金")
		fmt.Println("  max-position-pct      单只股票上限")
		return
	}

	key := args[0]
	value := args[1]

	switch key {
	case "default-target-yield":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			a.configMgr.Config.DefaultTargetYield = v
		}
	case "min-dividend-years":
		if v, err := strconv.Atoi(value); err == nil {
			a.configMgr.Config.MinDividendYears = v
		}
	case "cheap-discount":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			a.configMgr.Config.CheapDiscount = v / 100
		}
	case "expensive-premium":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			a.configMgr.Config.ExpensivePremium = v / 100
		}
	case "available-fund":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			a.configMgr.Config.Fund.AvailableFund = v
		}
	case "max-position-pct":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			a.configMgr.Config.Fund.MaxPositionPct = v
		}
	default:
		fmt.Printf("未知配置键: %s\n", key)
		return
	}

	if err := a.configMgr.Save(); err != nil {
		fmt.Printf("保存配置失败: %v\n", err)
		return
	}

	fmt.Printf("已设置 %s = %s\n", key, value)
}
