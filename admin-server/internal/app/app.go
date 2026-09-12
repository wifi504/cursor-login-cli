package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wifi504/cursor-login-cli/admin-server/internal/config"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/db"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/httpapi"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/store"
	"github.com/wifi504/logo"
)

var (
	scopeApp  = logo.RegisterScope("app")
	logServer = scopeApp.RegisterModule("server")
	logHTTP   = scopeApp.RegisterModule("http")
)

func Run() error {
	cfg := config.Parse()
	if err := cfg.Validate(); err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	exeDir := dataDirForRuntime(exe)
	logDir := filepath.Join(exeDir, "logs")

	if err := logo.Init(logo.Config{
		Level:     "debug",
		Dir:       logDir,
		MaxSizeMB: 20,
		Compress:  true,
		Stdout:    true,
		Scopes: map[string]logo.ScopeConfig{
			"app": {
				Level: "debug",
				Modules: map[string]logo.ModuleConfig{
					"server": {Level: "debug"},
					"http":   {Level: "debug"},
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}
	defer logo.Close()

	sqlDB, dbPath, err := db.OpenBesideBinary()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	logServer.Info("数据库已连接：%s", dbPath)

	st := store.New(sqlDB)
	entry, _, err := st.EnsureBootstrap()
	if err != nil {
		return err
	}
	initialized, _ := st.IsInitialized()

	listenHost := "http://127.0.0.1" + normalizeListen(cfg.Listen)

	if cfg.Dev {
		logServer.Warn("======== Cursor Login Server 开发模式 (--dev) ========")
		logServer.Warn("请只打开 Cursor Login Admin 前端: http://localhost:5173/")
		logServer.Warn("管理 API 固定走: %s%s/api/*", listenHost, httpapi.DevEntry)
		logServer.Warn("不要用浏览器直接打开 :8080 上的页面做日常开发")
	} else if !initialized {
		logServer.Warn("Cursor Login Server 等待初始化，请打开：%s%s/", listenHost, entry)
	} else {
		logServer.Info("Cursor Login Admin 请打开：%s%s/", listenHost, entry)
		if base, _ := st.PublicBaseURL(); base != "" {
			logServer.Info("对外访问地址: %s", base)
		}
	}

	srv := &httpapi.Server{
		Store:      st,
		ExeDir:     exeDir,
		Log:        logHTTP,
		AdminEntry: entry,
		DevMode:    cfg.Dev,
	}
	engine := httpapi.NewRouter(srv)

	httpServer := &http.Server{
		Addr:              cfg.Listen,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}
	logServer.Info("Cursor Login Server 监听 %s （根路径 / 故意返回 403）", cfg.Listen)
	return httpServer.ListenAndServe()
}

func dataDirForRuntime(exe string) string {
	dir := filepath.Dir(exe)
	if strings.Contains(filepath.ToSlash(dir), "/go-build") || strings.Contains(dir, `\go-build`) {
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
	}
	return dir
}

func normalizeListen(addr string) string {
	if len(addr) > 0 && addr[0] == ':' {
		return addr
	}
	return addr
}
