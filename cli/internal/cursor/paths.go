package cursor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Paths Cursor 用户数据路径。
type Paths struct {
	Root    string
	Storage string // storage.json
	DB      string // state.vscdb
}

func ResolvePaths() (*Paths, error) {
	var root string
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return nil, fmt.Errorf("未设置 APPDATA")
		}
		root = filepath.Join(appData, "Cursor")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(home, "Library", "Application Support", "Cursor")
	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(home, ".config", "Cursor")
	default:
		return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	gs := filepath.Join(root, "User", "globalStorage")
	return &Paths{
		Root:    root,
		Storage: filepath.Join(gs, "storage.json"),
		DB:      filepath.Join(gs, "state.vscdb"),
	}, nil
}

// EnsureWritable 检查 state.vscdb 存在且可打开写入。
func (p *Paths) EnsureWritable() error {
	st, err := os.Stat(p.DB)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("未找到 Cursor 本地数据，请至少运行过一次 Cursor 客户端")
		}
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("state.vscdb 路径异常")
	}
	f, err := os.OpenFile(p.DB, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("无法写入 state.vscdb: %w", err)
	}
	_ = f.Close()
	return nil
}
