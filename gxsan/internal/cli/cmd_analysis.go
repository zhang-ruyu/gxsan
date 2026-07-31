package cli

import (
	"fmt"
	"strconv"

	"github.com/user/gxsan/internal/report"
)

// hasRefreshFlag 检查参数中是否含 --refresh
func hasRefreshFlag(args []string) bool {
	for _, a := range args {
		if a == "--refresh" {
			return true
		}
	}
	return false
}

// handleAnalyze 处理分析（支持 --refresh 强制重新抓取）
func (a *App) handleAnalyze(args []string) {
	forceRefresh := hasRefreshFlag(args)
	if forceRefresh {
		fmt.Println("正在强制刷新数据并分析...")
	} else {
		fmt.Println("正在获取数据并分析...")
	}

	reportStr, err := a.reporter.GenerateAnalysisReport(forceRefresh)
	if err != nil {
		fmt.Printf("分析失败: %v\n", err)
		return
	}

	fmt.Println(reportStr)
}

// handleDetail 处理详情（支持 --refresh 强制重新抓取）
func (a *App) handleDetail(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: gxsan detail <代码> [--refresh]")
		return
	}

	code := args[0]
	forceRefresh := hasRefreshFlag(args)
	fmt.Printf("正在获取 %s 数据...\n", code)

	d, err := a.reporter.GenerateStockDetail(code, forceRefresh)
	if err != nil {
		fmt.Printf("获取详情失败: %v\n", err)
		return
	}

	fmt.Print(report.FormatStockDetailText(d))
}

// handleCalendar 处理股利日历
func (a *App) handleCalendar(args []string) {
	days := 30

	for i := 0; i < len(args); i++ {
		if args[i] == "--days" && i+1 < len(args) {
			if d, err := strconv.Atoi(args[i+1]); err == nil {
				days = d
			}
			i++
		}
	}

	fmt.Println("正在获取股利日历...")

	reportStr, err := a.reporter.GenerateCalendar(days)
	if err != nil {
		fmt.Printf("获取日历失败: %v\n", err)
		return
	}

	fmt.Println(reportStr)
}
