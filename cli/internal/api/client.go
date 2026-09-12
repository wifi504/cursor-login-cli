package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Base   string
	Client *http.Client
}

func New(base string) *Client {
	return &Client{
		Base: strings.TrimRight(strings.TrimSpace(base), "/"),
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type PreviewResp struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Remaining int    `json:"remaining"`
}

type RedeemResp struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
	Remaining   int    `json:"remaining"`
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func (c *Client) Health() error {
	resp, err := c.Client.Get(c.Base + "/api/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	_ = body
	return nil
}

func (c *Client) Preview(code string) (*PreviewResp, error) {
	var out PreviewResp
	if err := c.postJSON("/api/claim/preview", map[string]string{"code": code}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Redeem(code string) (*RedeemResp, error) {
	var out RedeemResp
	if err := c.postJSON("/api/claim", map[string]string{"code": code}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) postJSON(path string, body any, dest any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := c.Client.Post(c.Base+path, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var er struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(data, &er)
		msg := strings.TrimSpace(er.Error)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &APIError{Status: resp.StatusCode, Message: msg}
	}
	if dest != nil {
		if err := json.Unmarshal(data, dest); err != nil {
			return err
		}
	}
	return nil
}

// ReasonAfterColon 从错误中提取冒号后原因，不展示 URL。
func ReasonAfterColon(err error) string {
	if err == nil {
		return "未知错误"
	}
	msg := err.Error()
	// 去掉可能夹带的 URL
	msg = stripURLs(msg)
	if i := strings.LastIndex(msg, ": "); i >= 0 && i+2 < len(msg) {
		return strings.TrimSpace(msg[i+2:])
	}
	if i := strings.LastIndex(msg, ":"); i >= 0 && i+1 < len(msg) {
		return strings.TrimSpace(msg[i+1:])
	}
	return strings.TrimSpace(msg)
}

func stripURLs(s string) string {
	out := s
	for _, p := range []string{"https://", "http://"} {
		for {
			i := strings.Index(strings.ToLower(out), p)
			if i < 0 {
				break
			}
			j := i + len(p)
			for j < len(out) && out[j] != ' ' && out[j] != '"' && out[j] != '\'' {
				j++
			}
			out = out[:i] + out[j:]
		}
	}
	return out
}
