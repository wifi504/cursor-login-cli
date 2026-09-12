package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/store"
)

func (s *Server) requireLogin(c *gin.Context) (string, bool) {
	user, err := s.currentUser(c)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return "", false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return "", false
	}
	return user, true
}

func (s *Server) handleListAccounts(c *gin.Context) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	list, err := s.Store.ListAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []store.Account{}
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type accountBody struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
	ClaimCode   string `json:"claim_code"`
	ClaimLimit  *int   `json:"claim_limit"`
}

func (s *Server) handleCreateAccount(c *gin.Context) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	var req accountBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	a, err := s.Store.CreateAccount(store.AccountInput{
		Name:        req.Name,
		Email:       req.Email,
		AccessToken: req.AccessToken,
		ClaimCode:   req.ClaimCode,
		ClaimLimit:  req.ClaimLimit,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (s *Server) handleGetAccount(c *gin.Context, id int64) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	a, err := s.Store.GetAccount(id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (s *Server) handleUpdateAccount(c *gin.Context, id int64) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	var req accountBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	a, err := s.Store.UpdateAccount(id, store.AccountInput{
		Name:        req.Name,
		Email:       req.Email,
		AccessToken: req.AccessToken,
		ClaimCode:   req.ClaimCode,
	})
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

type claimLimitBody struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (s *Server) handleUpdateClaimLimit(c *gin.Context, id int64) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	var req claimLimitBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	a, err := s.Store.UpdateClaimLimit(id, req.From, req.To)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

type resetClaimedBody struct {
	From int `json:"from"`
}

func (s *Server) handleResetClaimed(c *gin.Context, id int64) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	var req resetClaimedBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	a, err := s.Store.ResetClaimedCount(id, req.From)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (s *Server) handleDeleteAccount(c *gin.Context, id int64) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	err := s.Store.DeleteAccount(id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "账号不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// routeAccounts 解析 /accounts 与子路径。
func (s *Server) routeAccounts(c *gin.Context, apiPath string) bool {
	if apiPath == "/accounts" {
		switch c.Request.Method {
		case http.MethodGet:
			s.handleListAccounts(c)
		case http.MethodPost:
			s.handleCreateAccount(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "方法不允许"})
		}
		return true
	}
	if !strings.HasPrefix(apiPath, "/accounts/") {
		return false
	}
	rest := strings.TrimPrefix(apiPath, "/accounts/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
		return true
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账号 ID"})
		return true
	}
	if len(parts) == 1 {
		switch c.Request.Method {
		case http.MethodGet:
			s.handleGetAccount(c, id)
		case http.MethodPut:
			s.handleUpdateAccount(c, id)
		case http.MethodDelete:
			s.handleDeleteAccount(c, id)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "方法不允许"})
		}
		return true
	}
	if len(parts) == 2 && parts[1] == "claim-limit" && c.Request.Method == http.MethodPut {
		s.handleUpdateClaimLimit(c, id)
		return true
	}
	if len(parts) == 2 && parts[1] == "reset-claimed" && c.Request.Method == http.MethodPost {
		s.handleResetClaimed(c, id)
		return true
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
	return true
}
