package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/gxsan/internal/config"
	"github.com/user/gxsan/internal/data"
	"github.com/user/gxsan/internal/report"
)

// App 应用程序
type App struct {
	configMgr *config.Manager
	fetcher   *data.EastMoneyFetcher
	cache     *data.Cache
	reporter  *report.Reporter
}

// NewApp 创建应用程序
func NewApp() *App {
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	fetcher := data.NewEastMoneyFetcher()
	cache := data.NewCache(configMgr.DataPath)
	reporter := report.NewReporter(configMgr.Config, fetcher, cache)

	return &App{
		configMgr: configMgr,
		fetcher:   fetcher,
		cache:     cache,
		reporter:  reporter,
	}
}

// Run 运行
func (a *App) Run(args []string) {
	if len(args) < 2 {
		a.showHelp()
		return
	}

	command := args[1]

	switch command {
	case "add":
		a.handleAdd(args[2:])
	case "list":
		a.handleList()
	case "remove":
		a.handleRemove(args[2:])
	case "search":
		a.handleSearch(args[2:])
	case "grid":
		a.handleGrid(args[2:])
	case "portfolio":
		a.handlePortfolio(args[2:])
	case "fund":
		a.handleFund(args[2:])
	case "analyze":
		a.handleAnalyze(args[2:])
	case "detail":
		a.handleDetail(args[2:])
	case "calendar":
		a.handleCalendar(args[2:])
	case "pension":
		a.handlePension(args[2:])
	case "compare":
		a.handleCompare(args[2:])
	case "yoc":
		a.handleYoC(args[2:])
	case "config":
		a.handleConfig(args[2:])
	case "version":
		a.handleVersion()
	case "help":
		a.showHelp()
	default:
		fmt.Printf("未知命令: %s\n", command)
		a.showHelp()
	}
}

// showHelp 显示帮助
func (a *App) showHelp() {
	fmt.Print(`股息三 (GxSan) - A股股息分析与投资决策辅助工具

用法:
  gxsan [命令] [参数]

命令:
  add <代码> <名称> [--target-yield <目标股息率>]
      添加股票到监控列表

  list
      查看所有监控股票

  remove <代码>
      从列表删除股票

  search <关键词>
      搜索股票

  grid set <代码> --buy <档位> <股息率> <金额>
  grid set <代码> --sell <档位> <股息率> <比例>
  grid show <代码>
      管理网格策略

  portfolio add <代码> --shares <股数> --cost <成本价>
  portfolio update <代码> --shares <股数>
  portfolio list
  portfolio remove <代码>
      管理投资组合

  fund set <金额>
  fund recommend
      资金管理

  analyze [--refresh]
      运行分析并生成报告（--refresh 强制重新抓取）

  detail <代码> [--refresh]
      查看单只股票详情（--refresh 强制重新抓取）

  calendar [--days <天数>]
      显示股利日历

  pension [--monthly <目标月分红>] [--invest <月定投>] [--years <年数>] [--rate <年化%>]
      养老现金流测算（目标倒推所需本金 + 定投复利模拟）

  compare <代码A> <代码B>
      两只股票多维对比（股息率/估值/分红稳定性）

  yoc
      持仓成本股息率(YoC)

  config show
  config set <键> <值>
      配置管理

  version
      显示版本信息

  help
      显示此帮助信息
`)
}

// handleVersion 处理版本
func (a *App) handleVersion() {
	fmt.Println("股息三 (GxSan) v1.0.0")
	fmt.Println("A股股息分析与投资决策辅助工具")
}

// ParseAmount 解析金额字符串
func ParseAmount(s string) float64 {
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "元", "")

	var amount float64
	fmt.Sscanf(s, "%f", &amount)
	return amount
}
