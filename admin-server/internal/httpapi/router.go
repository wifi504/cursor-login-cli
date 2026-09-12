package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/install"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/store"
	"github.com/wifi504/cursor-login-cli/admin-server/releases"
	"github.com/wifi504/cursor-login-cli/admin-server/web"
	"github.com/wifi504/logo"
)

const (
	cookieName   = "cursor_login_session"
	sessionTTL   = 7 * 24 * time.Hour
	// DevEntry 仅 --dev 时启用，供 Vite 固定代理，与真实安全入口无关
	DevEntry = "/__dev__"
)

type Server struct {
	Store      *store.Store
	ExeDir     string
	Log        *logo.Logger
	AdminEntry string // 当前管理端安全入口，初始化完成后会更新
	DevMode    bool
}

func NewRouter(s *Server) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/api/health", s.handleHealth)

	r.GET("/install.sh", s.handleInstallSH)
	r.GET("/uninstall.sh", s.handleUninstallSH)
	r.GET("/install.ps1", s.handleInstallPS1)
	r.GET("/uninstall.ps1", s.handleUninstallPS1)
	r.GET("/download/:name", s.handleDownload)

	// 其余路径：根路径禁止；管理端挂在安全入口下
	r.NoRoute(s.handleNoRoute)

	return r
}

func (s *Server) handleHealth(c *gin.Context) {
	ok, _ := s.Store.IsInitialized()
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"service":     "Cursor Login Server",
		"initialized": ok,
	})
}

func (s *Server) requirePublicBase(c *gin.Context) (string, bool) {
	base, err := s.Store.PublicBaseURL()
	if err != nil || base == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "尚未配置对外访问地址"})
		return "", false
	}
	return base, true
}

func (s *Server) handleInstallSH(c *gin.Context) {
	base, ok := s.requirePublicBase(c)
	if !ok {
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, install.InstallSH(base))
}

func (s *Server) handleUninstallSH(c *gin.Context) {
	base, ok := s.requirePublicBase(c)
	if !ok {
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, install.UninstallSH(base))
}

func (s *Server) handleInstallPS1(c *gin.Context) {
	base, ok := s.requirePublicBase(c)
	if !ok {
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, install.InstallPS1(base))
}

func (s *Server) handleUninstallPS1(c *gin.Context) {
	base, ok := s.requirePublicBase(c)
	if !ok {
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, install.UninstallPS1(base))
}

func (s *Server) handleDownload(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	if name == "." || name == "" || strings.Contains(name, "..") || name == ".gitkeep" {
		c.Status(http.StatusBadRequest)
		return
	}

	// 优先读嵌入产物；磁盘旁路仅便于本地临时覆盖（无需重编服务端）
	var data []byte
	diskPath := filepath.Join(s.ExeDir, "releases", "bin", name)
	if b, err := os.ReadFile(diskPath); err == nil && len(b) > 0 {
		data = b
	} else if b, err := releases.Read(name); err == nil && len(b) > 0 {
		data = b
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该平台的 CLI 安装包，请确认发版时已嵌入或放入 releases/bin/"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+name)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.Itoa(len(data)))
	_, _ = c.Writer.Write(data)
}

func (s *Server) handleNoRoute(c *gin.Context) {
	path := c.Request.URL.Path

	// 开发模式固定入口：始终可用，方便 Vite 只打开 localhost:5173
	if s.DevMode {
		if path == DevEntry {
			c.Redirect(http.StatusFound, DevEntry+"/")
			return
		}
		if strings.HasPrefix(path, DevEntry+"/") {
			rel := strings.TrimPrefix(path, DevEntry)
			if rel == "" {
				rel = "/"
			}
			s.serveAdmin(c, rel)
			return
		}
	}

	entry, err := s.Store.AdminEntry()
	if err != nil || entry == "" {
		c.Status(http.StatusForbidden)
		return
	}

	// 无尾斜杠时重定向到带尾斜杠，否则 base:'./' 的静态资源会解析到错误路径（如 /入口/../assets）
	if path == entry {
		c.Redirect(http.StatusFound, entry+"/")
		return
	}

	// 精确匹配安全入口子路径
	if strings.HasPrefix(path, entry+"/") {
		rel := strings.TrimPrefix(path, entry)
		if rel == "" {
			rel = "/"
		}
		s.serveAdmin(c, rel)
		return
	}

	c.Status(http.StatusForbidden)
}

func (s *Server) serveAdmin(c *gin.Context, rel string) {
	// 管理端 JSON API：{入口}/api/*
	if strings.HasPrefix(rel, "/api/") || rel == "/api" {
		s.serveAdminAPI(c, strings.TrimPrefix(rel, "/api"))
		return
	}

	dist, err := web.FS()
	if err != nil {
		c.String(http.StatusInternalServerError, "前端资源不可用")
		return
	}

	filePath := strings.TrimPrefix(rel, "/")
	if filePath == "" || strings.HasSuffix(rel, "/") {
		filePath = "index.html"
	}

	data, err := fs.ReadFile(dist, filePath)
	if err != nil {
		data, err = fs.ReadFile(dist, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "缺少 index.html")
			return
		}
		filePath = "index.html"
	}

	ctype := contentType(filePath)
	c.Data(http.StatusOK, ctype, data)
}

func contentType(name string) string {
	switch {
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(name, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(name, ".json"):
		return "application/json"
	case strings.HasSuffix(name, ".png"):
		return "image/png"
	case strings.HasSuffix(name, ".ico"):
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

func (s *Server) serveAdminAPI(c *gin.Context, apiPath string) {
	if apiPath == "" {
		apiPath = "/"
	}
	switch {
	case c.Request.Method == http.MethodGet && apiPath == "/status":
		s.handleAdminStatus(c)
	case c.Request.Method == http.MethodPost && apiPath == "/bootstrap":
		s.handleBootstrap(c)
	case c.Request.Method == http.MethodPost && apiPath == "/login":
		s.handleLogin(c)
	case c.Request.Method == http.MethodPost && apiPath == "/logout":
		s.handleLogout(c)
	case c.Request.Method == http.MethodGet && apiPath == "/me":
		s.handleMe(c)
	case c.Request.Method == http.MethodGet && apiPath == "/install-commands":
		s.handleInstallCommands(c)
	case c.Request.Method == http.MethodPut && apiPath == "/settings/entry":
		s.handleUpdateEntry(c)
	case c.Request.Method == http.MethodPut && apiPath == "/settings/account":
		s.handleUpdateAccountSettings(c)
	case c.Request.Method == http.MethodGet && apiPath == "/cursor/auth/poll":
		s.handleCursorAuthPoll(c)
	default:
		if s.routeAccounts(c, apiPath) {
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
	}
}

func (s *Server) handleAdminStatus(c *gin.Context) {
	ok, err := s.Store.IsInitialized()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	entry, _ := s.Store.AdminEntry()
	base, _ := s.Store.PublicBaseURL()
	c.JSON(http.StatusOK, gin.H{
		"initialized":     ok,
		"admin_entry":     entry,
		"public_base_url": base,
	})
}

type bootstrapReq struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	AdminEntry    string `json:"admin_entry"`
	PublicBaseURL string `json:"public_base_url"`
}

func (s *Server) handleBootstrap(c *gin.Context) {
	var req bootstrapReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	if err := s.Store.CompleteBootstrap(store.BootstrapInput{
		Username:      req.Username,
		Password:      req.Password,
		AdminEntry:    req.AdminEntry,
		PublicBaseURL: req.PublicBaseURL,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, _ := s.Store.AdminEntry()
	s.AdminEntry = entry
	s.Log.Info("初始化完成，Cursor Login Admin 安全入口现为 %s", entry)
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"admin_entry": entry,
	})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(c *gin.Context) {
	ok, err := s.Store.IsInitialized()
	if err != nil || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "系统尚未初始化"})
		return
	}
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	pass, err := s.Store.VerifyAdmin(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !pass {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, expires, err := s.Store.CreateSession(req.Username, sessionTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.SetCookie(cookieName, token, int(sessionTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true, "username": req.Username, "expires_at": expires.Format(time.RFC3339)})
}

func (s *Server) handleLogout(c *gin.Context) {
	token, _ := c.Cookie(cookieName)
	if token != "" {
		_ = s.Store.DeleteSession(token)
	}
	c.SetCookie(cookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) currentUser(c *gin.Context) (string, error) {
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		return "", store.ErrNotFound
	}
	return s.Store.GetSession(token)
}

func (s *Server) handleMe(c *gin.Context) {
	user, err := s.currentUser(c)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": user})
}

func (s *Server) handleInstallCommands(c *gin.Context) {
	if _, err := s.currentUser(c); errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base, ok := s.requirePublicBase(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, install.Commands(base))
}

type updateEntryReq struct {
	AdminEntry      string `json:"admin_entry"`
	CurrentPassword string `json:"current_password"`
}

func (s *Server) handleUpdateEntry(c *gin.Context) {
	if _, err := s.currentUser(c); errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var req updateEntryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	if err := s.Store.UpdateAdminEntry(req.AdminEntry, req.CurrentPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, _ := s.Store.AdminEntry()
	s.AdminEntry = entry
	s.Log.Info("Cursor Login Admin 安全入口已更新为 %s", entry)
	c.JSON(http.StatusOK, gin.H{"ok": true, "admin_entry": entry})
}

type updateAccountReq struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	CurrentPassword string `json:"current_password"`
}

func (s *Server) handleUpdateAccountSettings(c *gin.Context) {
	if _, err := s.currentUser(c); errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var req updateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	relogin, err := s.Store.UpdateAdminAccount(req.Username, req.Password, req.CurrentPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if relogin {
		c.SetCookie(cookieName, "", -1, "/", "", false, true)
		s.Log.Info("Cursor Login Admin 账号已更新，需重新登录")
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "relogin": relogin})
}
