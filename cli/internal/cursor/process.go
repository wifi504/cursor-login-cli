package cursor

import (
	"os/exec"
	"runtime"
	"strings"
)

// IsRunning 检测 Cursor 客户端是否在运行（不杀进程）。
func IsRunning() bool {
	switch runtime.GOOS {
	case "windows":
		return windowsCursorRunning()
	default:
		return unixCursorRunning()
	}
}

func windowsCursorRunning() bool {
	out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		low := strings.ToLower(line)
		if !strings.Contains(low, "cursor.exe") && !strings.Contains(low, "\"cursor\"") {
			continue
		}
		// 排除本工具相关进程名
		if strings.Contains(low, "cursor-login") || strings.Contains(low, "cursor-pool") {
			continue
		}
		return true
	}
	return false
}

func unixCursorRunning() bool {
	out, err := exec.Command("ps", "-ax", "-o", "comm=").Output()
	if err != nil {
		out, err = exec.Command("ps", "-A", "-o", "comm=").Output()
		if err != nil {
			return false
		}
	}
	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(line)
		base := name
		if i := strings.LastIndex(name, "/"); i >= 0 {
			base = name[i+1:]
		}
		low := strings.ToLower(base)
		if low == "cursor" || strings.HasPrefix(low, "cursor ") {
			return true
		}
		// macOS 常见 Cursor Helper
		if strings.Contains(low, "cursor") && !strings.Contains(low, "cursor-login") && !strings.Contains(low, "cursor-pool") {
			if low == "cursor" || strings.HasPrefix(low, "cursor helper") {
				return true
			}
		}
	}
	return false
}
