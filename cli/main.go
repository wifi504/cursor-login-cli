package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wifi504/cursor-login-cli/cli/internal/api"
	"github.com/wifi504/cursor-login-cli/cli/internal/cursor"
)

// version 由 -ldflags "-X main.version=..." 注入；默认 dev。
var version = "dev"

const envAPI = "CURSOR_LOGIN_API"

func main() {
	os.Exit(run(os.Args[1:]))
}

func banner() string {
	return fmt.Sprintf("Cursor Login CLI (%s)", version)
}

func printHelp() {
	fmt.Println(banner())
	fmt.Println("使用方式：cursor-login <上号码>")
	fmt.Println("请至少运行过一次 Cursor 客户端！")
}

func isHelp(arg string) bool {
	switch strings.ToLower(strings.TrimSpace(arg)) {
	case "help", "-h", "-help", "--help":
		return true
	default:
		return false
	}
}

func run(args []string) int {
	if len(args) == 0 || isHelp(args[0]) {
		printHelp()
		return 0
	}
	if len(args) > 1 {
		printHelp()
		return 1
	}
	code := strings.TrimSpace(args[0])
	if code == "" || isHelp(code) {
		printHelp()
		return 0
	}

	base := strings.TrimSpace(os.Getenv(envAPI))
	if base == "" {
		fmt.Println(banner())
		fmt.Println("服务端连接失败，请重新安装后再次尝试！")
		fmt.Println("尚未完成安装配置")
		return 1
	}

	client := api.New(base)
	if err := client.Health(); err != nil {
		fmt.Println(banner())
		fmt.Println("服务端连接失败，请重新安装后再次尝试！")
		fmt.Println(api.ReasonAfterColon(err))
		return 1
	}

	prev, err := client.Preview(code)
	if err != nil {
		var ae *api.APIError
		if errors.As(err, &ae) && ae.Status == 404 {
			fmt.Println(banner())
			fmt.Println(ae.Message)
			return 1
		}
		fmt.Println(banner())
		fmt.Println("服务端连接失败，请重新安装后再次尝试！")
		fmt.Println(api.ReasonAfterColon(err))
		return 1
	}

	fmt.Println(banner())
	fmt.Printf("获取账号“%s”成功！\n", prev.Email)
	fmt.Printf("信息：%s\n", prev.Name)

	if prev.Remaining <= 0 {
		fmt.Println("核销次数已用尽，请联系管理员！")
		return 1
	}

	fmt.Printf("剩余核销次数：%d\n", prev.Remaining)

	if cursor.IsRunning() {
		fmt.Println("请先关闭运行中的 Cursor 后再尝试，注意保存未保存的工作！")
		return 1
	}

	paths, err := cursor.ResolvePaths()
	if err != nil {
		fmt.Printf("无法定位 Cursor 数据目录：%v\n", err)
		return 1
	}
	if err := paths.EnsureWritable(); err != nil {
		fmt.Println(err.Error())
		return 1
	}

	fmt.Print("确认要核销此账号吗？(y/n): ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	ans := strings.TrimSpace(strings.ToLower(line))
	if ans != "y" {
		return 0
	}

	fmt.Println("正在登录账号...")
	redeem, err := client.Redeem(code)
	if err != nil {
		var ae *api.APIError
		if errors.As(err, &ae) {
			fmt.Println(ae.Message)
			return 1
		}
		fmt.Println("服务端连接失败，请重新安装后再次尝试！")
		fmt.Println(api.ReasonAfterColon(err))
		return 1
	}

	if err := cursor.ResetTelemetry(paths); err != nil {
		fmt.Printf("写入本地设备身份失败：%v\n", err)
		fmt.Println("次数可能已扣减，请联系管理员处理。")
		return 1
	}
	if err := cursor.WriteAuth(paths, redeem.Email, redeem.AccessToken); err != nil {
		fmt.Printf("写入本地登录态失败：%v\n", err)
		fmt.Println("次数可能已扣减，请联系管理员处理。")
		return 1
	}

	fmt.Println("登录成功！")
	return 0
}
