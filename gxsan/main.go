package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/user/gxsan/internal/cli"
)

//go:embed all:frontend/dist
var assets embed.FS

// cliCommands CLI 子命令白名单：双击 exe 走 GUI，带这些子命令时走 CLI
var cliCommands = map[string]bool{
	"add": true, "list": true, "remove": true, "search": true,
	"grid": true, "portfolio": true, "fund": true, "analyze": true,
	"detail": true, "calendar": true, "pension": true, "compare": true,
	"yoc": true, "config": true, "version": true, "help": true,
}

func main() {
	// 双入口：有 CLI 子命令走命令行，否则启动 GUI
	if len(os.Args) >= 2 && cliCommands[os.Args[1]] {
		app := cli.NewApp()
		app.Run(os.Args)
		return
	}
	runGUI()
}

func runGUI() {
	app := NewApp()

	// WebView2 固定版运行时：如果 exe 同目录下有 webview2/ 目录则使用，否则回退系统 WebView2
	winOpts := &windows.Options{}
	if _, err := os.Stat("webview2"); err == nil {
		winOpts.WebviewBrowserPath = "webview2"
	}

	err := wails.Run(&options.App{
		Title:     "股息三 v1.0.0",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 600,
		Frameless: true, // 无边框：隐藏原生标题栏，改用前端自定义标题栏（左上角 最小化/全屏/关闭）
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 250, A: 1},
		Windows:          winOpts,
		OnStartup:        app.startup,
		// 关闭即退出：避免隐藏进程残留导致多次双击堆积、争抢行情接口拖慢启动。
		// 若日后需要最小化到托盘，再改为单实例锁 + 托盘逻辑。
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			return false
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
