package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/store"
)

type claimCodeBody struct {
	Code string `json:"code"`
}

const (
	msgClaimCodeInvalid = "上号码错误或不存在，请核对后再尝试！"
	msgClaimExhausted   = "核销次数已用尽，请联系管理员！"
	msgTokenNotReady    = "账号未就绪，请联系管理员"
)

func (s *Server) handleClaimPreview(c *gin.Context) {
	var req claimCodeBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	code := strings.TrimSpace(req.Code)
	ip := c.ClientIP()

	prev, err := s.Store.PreviewByClaimCode(code)
	if errors.Is(err, store.ErrNotFound) {
		s.Log.Info("claim_preview code=%s result=not_found ip=%s", code, ip)
		c.JSON(http.StatusNotFound, gin.H{"error": msgClaimCodeInvalid})
		return
	}
	if errors.Is(err, store.ErrTokenNotReady) {
		s.Log.Info("claim_preview code=%s result=not_ready ip=%s", code, ip)
		c.JSON(http.StatusBadRequest, gin.H{"error": msgTokenNotReady})
		return
	}
	if err != nil {
		s.Log.Error("claim_preview code=%s result=error ip=%s err=%v", code, ip, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.Log.Info("claim_preview code=%s result=ok account_id=%d email=%s remaining=%d ip=%s",
		code, prev.ID, prev.Email, prev.Remaining, ip)
	c.JSON(http.StatusOK, gin.H{
		"email":     prev.Email,
		"name":      prev.Name,
		"remaining": prev.Remaining,
	})
}

func (s *Server) handleClaimRedeem(c *gin.Context) {
	var req claimCodeBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体 JSON 无效"})
		return
	}
	code := strings.TrimSpace(req.Code)
	ip := c.ClientIP()

	out, err := s.Store.RedeemByClaimCode(code)
	if errors.Is(err, store.ErrNotFound) {
		s.Log.Info("claim_redeem code=%s result=not_found ip=%s", code, ip)
		c.JSON(http.StatusNotFound, gin.H{"error": msgClaimCodeInvalid})
		return
	}
	if errors.Is(err, store.ErrClaimExhausted) {
		s.Log.Info("claim_redeem code=%s result=exhausted ip=%s", code, ip)
		c.JSON(http.StatusConflict, gin.H{"error": msgClaimExhausted})
		return
	}
	if errors.Is(err, store.ErrTokenNotReady) {
		s.Log.Info("claim_redeem code=%s result=token_not_ready ip=%s", code, ip)
		c.JSON(http.StatusBadRequest, gin.H{"error": msgTokenNotReady})
		return
	}
	if err != nil {
		s.Log.Error("claim_redeem code=%s result=error ip=%s err=%v", code, ip, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.Log.Info("claim_redeem code=%s result=ok account_id=%d email=%s remaining=%d ip=%s",
		code, out.ID, out.Email, out.Remaining, ip)
	c.JSON(http.StatusOK, gin.H{
		"email":         out.Email,
		"name":          out.Name,
		"access_token":  out.AccessToken,
		"remaining":     out.Remaining,
	})
}
