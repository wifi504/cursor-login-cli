package main

import (
	"fmt"
	"os"

	"github.com/wifi504/cursor-login-cli/admin-server/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cursor Login Server 启动失败: %v\n", err)
		os.Exit(1)
	}
}
