package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const envAPI = "CURSOR_LOGIN_API"

func main() {
	api := strings.TrimSpace(os.Getenv(envAPI))
	if api == "" {
		fmt.Fprintf(os.Stderr, "Cursor Login CLI：无法连接服务器——尚未完成安装配置。\n")
		fmt.Fprintf(os.Stderr, "请先使用管理员提供的一键安装命令安装 Cursor Login CLI。\n")
		os.Exit(1)
	}
	api = strings.TrimRight(api, "/")

	client := &http.Client{Timeout: 8 * time.Second}
	url := api + "/api/health"
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cursor Login CLI：无法连接服务器：%v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Cursor Login CLI：无法连接服务器：HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	fmt.Println("Cursor Login CLI：已连接到 Cursor Login Server")
	if ok, _ := payload["ok"].(bool); ok {
		fmt.Println("健康检查：通过")
	}
	if init, ok := payload["initialized"].(bool); ok {
		fmt.Printf("服务已初始化：%v\n", init)
	}
}
