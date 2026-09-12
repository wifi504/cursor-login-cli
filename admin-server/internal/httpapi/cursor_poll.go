package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

const cursorPollUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Cursor/0.48.6 Chrome/132.0.6834.210 Electron/34.3.4 Safari/537.36"

// handleCursorAuthPoll 代理 Cursor auth/poll，供管理端获取长期 AccessToken。
func (s *Server) handleCursorAuthPoll(c *gin.Context) {
	if _, ok := s.requireLogin(c); !ok {
		return
	}
	uuid := c.Query("uuid")
	verifier := c.Query("verifier")
	if uuid == "" || verifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 uuid 或 verifier"})
		return
	}

	q := url.Values{}
	q.Set("uuid", uuid)
	q.Set("verifier", verifier)
	pollURL := "https://api2.cursor.sh/auth/poll?" + q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, pollURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "构造请求失败"})
		return
	}
	req.Header.Set("User-Agent", cursorPollUA)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "网络请求失败"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "上游 HTTP " + resp.Status, "body": string(body)})
		return
	}
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		c.Data(http.StatusOK, "application/json", body)
		return
	}
	c.JSON(http.StatusOK, data)
}
