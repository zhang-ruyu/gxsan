package cli

import (
	"fmt"
	"strconv"
)

// handleAdd 处理添加股票
func (a *App) handleAdd(args []string) {
	if len(args) < 2 {
		fmt.Println("用法: gxsan add <代码> <名称> [--target-yield <目标股息率>]")
		return
	}

	code := args[0]
	name := args[1]
	targetYield := a.configMgr.Config.DefaultTargetYield

	// 解析参数
	for i := 2; i < len(args); i++ {
		if args[i] == "--target-yield" && i+1 < len(args) {
			if y, err := strconv.ParseFloat(args[i+1], 64); err == nil {
				targetYield = y
			}
			i++
		}
	}

	if err := a.configMgr.AddStock(code, name, targetYield); err != nil {
		fmt.Printf("添加失败: %v\n", err)
		return
	}

	fmt.Printf("已添加股票: %s (%s), 目标股息率: %.2f%%\n", name, code, targetYield)
}

// handleList 处理列表
func (a *App) handleList() {
	if len(a.configMgr.Config.Watchlist) == 0 {
		fmt.Println("监控列表为空")
		return
	}

	fmt.Println("监控列表:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("%-8s %-10s %-10s\n", "代码", "名称", "目标股息率")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for _, stock := range a.configMgr.Config.Watchlist {
		fmt.Printf("%-8s %-10s %.2f%%\n", stock.Code, stock.Name, stock.TargetYield)
	}
}

// handleRemove 处理删除
func (a *App) handleRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: gxsan remove <代码>")
		return
	}

	code := args[0]
	if err := a.configMgr.RemoveStock(code); err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	fmt.Printf("已删除股票: %s\n", code)
}

// handleSearch 处理搜索
func (a *App) handleSearch(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: gxsan search <关键词>")
		return
	}

	keyword := args[0]
	stocks, err := a.fetcher.SearchStock(keyword)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	if len(stocks) == 0 {
		fmt.Printf("未找到匹配 '%s' 的股票\n", keyword)
		return
	}

	fmt.Printf("搜索结果 '%s':\n", keyword)
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("%-8s %-10s\n", "代码", "名称")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for _, stock := range stocks {
		fmt.Printf("%-8s %-10s\n", stock.Code, stock.Name)
	}
}
