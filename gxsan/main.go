package main

import (
	"context"
	"embed"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/user/gxsan/internal/cli"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.ico
var iconData []byte

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procLoadImage       = user32.NewProc("LoadImageW")
	procSendMessage     = user32.NewProc("SendMessageW")
	procFindWindow      = user32.NewProc("FindWindowW")
	procSetClassLongPtr = user32.NewProc("SetClassLongPtrW")
)

const (
	WM_SETICON      = 0x0080
	ICON_SMALL      = 0
	ICON_BIG        = 1
	GCL_HICON       = ^uintptr(13) // -14
	GCL_HICONSM     = ^uintptr(15) // -16
	LR_LOADFROMFILE = 0x0010
	IMAGE_ICON      = 1
)

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

	err := wails.Run(&options.App{
		Title:     "股息三 v1.0.0",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 250, A: 1},
		OnStartup:        app.startup,
		OnDomReady: func(ctx context.Context) {
			go func() {
				time.Sleep(1 * time.Second)
				setWindowIcon()
			}()
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			runtime.WindowHide(ctx)
			return true
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}

func setWindowIcon() {
	// 保存图标到临时文件
	tmpFile := os.TempDir() + "\\gxsan_icon.ico"
	if err := os.WriteFile(tmpFile, iconData, 0644); err != nil {
		return
	}
	defer os.Remove(tmpFile)

	// 通过窗口标题查找我们的窗口
	titlePtr, _ := syscall.UTF16PtrFromString("股息三 v1.0.0")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return
	}

	// 加载图标
	pathPtr, _ := syscall.UTF16PtrFromString(tmpFile)
	hIcon, _, _ := procLoadImage.Call(
		0,
		uintptr(unsafe.Pointer(pathPtr)),
		IMAGE_ICON,
		0,
		0,
		LR_LOADFROMFILE,
	)
	if hIcon == 0 {
		return
	}

	// 设置窗口类图标
	procSetClassLongPtr.Call(hwnd, GCL_HICON, hIcon)
	procSetClassLongPtr.Call(hwnd, GCL_HICONSM, hIcon)

	// 同时用SendMessage设置
	procSendMessage.Call(hwnd, WM_SETICON, ICON_BIG, hIcon)
	procSendMessage.Call(hwnd, WM_SETICON, ICON_SMALL, hIcon)
}
