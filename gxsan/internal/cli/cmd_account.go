package cli

import (
	"fmt"
)

// handleAccount 多账户体系管理
func (a *App) handleAccount(args []string) {
	if len(args) < 1 {
		fmt.Println("用法:")
		fmt.Println("  gxsan account list")
		fmt.Println("  gxsan account add <名称> <类型>   (类型: 养老/教育/港美股/娱乐/打新)")
		fmt.Println("  gxsan account remove <名称>")
		fmt.Println("  gxsan account assign <代码> <账户名>")
		return
	}

	switch args[0] {
	case "list":
		if len(a.configMgr.Config.Accounts) == 0 {
			fmt.Println("暂无账户，使用 account add 添加")
			return
		}
		fmt.Println("账户列表:")
		for _, acc := range a.configMgr.Config.Accounts {
			fmt.Printf("  %s (%s)\n", acc.Name, acc.Type)
		}

	case "add":
		if len(args) < 3 {
			fmt.Println("用法: gxsan account add <名称> <类型>")
			return
		}
		if err := a.configMgr.AddAccount(args[1], args[2]); err != nil {
			fmt.Printf("添加失败: %v\n", err)
			return
		}
		fmt.Printf("已添加账户: %s (%s)\n", args[1], args[2])

	case "remove":
		if len(args) < 2 {
			fmt.Println("用法: gxsan account remove <名称>")
			return
		}
		if err := a.configMgr.RemoveAccount(args[1]); err != nil {
			fmt.Printf("删除失败: %v\n", err)
			return
		}
		fmt.Printf("已删除账户: %s（相关持仓置为未分配）\n", args[1])

	case "assign":
		if len(args) < 3 {
			fmt.Println("用法: gxsan account assign <代码> <账户名>")
			return
		}
		if err := a.configMgr.SetHoldingAccount(args[1], args[2]); err != nil {
			fmt.Printf("分配失败: %v\n", err)
			return
		}
		fmt.Printf("已将持仓 %s 分配到账户 %s\n", args[1], args[2])

	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}
