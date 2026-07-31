package main

import (
	"os"

	"github.com/user/gxsan/internal/cli"
)

func main() {
	app := cli.NewApp()
	app.Run(os.Args)
}
