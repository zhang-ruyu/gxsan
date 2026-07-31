package cli

import (
	"fmt"
	"strconv"
)

// handleFund 处理资金管理
func (a *App) handleFund(args []string) {
	if len(args) < 1 {
		fmt.Println("用法:")
		fmt.Println("  gxsan fund set <金额>")
		fmt.Println("  gxsan fund recommend")
		return
	}

	switch args[0] {
	case "set":
		a.handleFundSet(args[1:])
	case "recommend":
		a.handleFundRecommend()
	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

// handleFundSet 处理设置资金
func (a *App) handleFundSet(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: gxsan fund set <金额>")
		return
	}

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		fmt.Printf("无效的金额: %s\n", args[0])
		return
	}

	if err := a.configMgr.SetFund(amount); err != nil {
		fmt.Printf("设置失败: %v\n", err)
		return
	}

	fmt.Printf("已设置可用资金: ¥%.0f\n", amount)
}

// handleFundRecommend 处理资金推荐
func (a *App) handleFundRecommend() {
	fmt.Println("正在分析...")
	// 资金推荐与分析共用同一份报告（含资金监控与操作推荐）
	a.handleAnalyze(nil)
}
