package cursor

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func normalizeToken(token string) string {
	token = strings.TrimSpace(token)
	if i := strings.Index(token, "%3A%3A"); i >= 0 {
		return token[i+6:]
	}
	if i := strings.Index(token, "::"); i >= 0 {
		return token[i+2:]
	}
	return token
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func generateIDs() (map[string]string, error) {
	dev, err := newUUID()
	if err != nil {
		return nil, err
	}
	raw64, err := randomHex(64)
	if err != nil {
		return nil, err
	}
	raw32, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	mac := sha512.Sum512([]byte(raw64))
	mid := sha256.Sum256([]byte(raw32))
	sqm, err := newUUID()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"telemetry.devDeviceId":  dev,
		"telemetry.macMachineId": hex.EncodeToString(mac[:]),
		"telemetry.machineId":    hex.EncodeToString(mid[:]),
		"telemetry.sqmId":        "{" + strings.ToUpper(sqm) + "}",
	}, nil
}

func upsertItem(db *sql.DB, key, value string) error {
	_, err := db.Exec(`INSERT INTO ItemTable (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		// 兼容无 UPSERT 的旧表结构
		res, err2 := db.Exec(`UPDATE ItemTable SET value = ? WHERE key = ?`, value, key)
		if err2 != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			_, err3 := db.Exec(`INSERT INTO ItemTable (key, value) VALUES (?, ?)`, key, value)
			return err3
		}
	}
	return nil
}

// ResetTelemetry 轻量重置 storage.json + state.vscdb 设备身份。
func ResetTelemetry(p *Paths) error {
	ids, err := generateIDs()
	if err != nil {
		return err
	}

	storage := map[string]any{}
	if raw, err := os.ReadFile(p.Storage); err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &storage)
		if storage == nil {
			storage = map[string]any{}
		}
	}
	for k, v := range ids {
		storage[k] = v
	}
	out, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.Storage), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(p.Storage, out, 0o644); err != nil {
		return fmt.Errorf("写入 storage.json 失败: %w", err)
	}

	db, err := sql.Open("sqlite", p.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS ItemTable (key TEXT PRIMARY KEY, value TEXT)`); err != nil {
		return err
	}
	for k, v := range ids {
		if err := upsertItem(db, k, v); err != nil {
			return err
		}
	}
	if err := upsertItem(db, "storage.serviceMachineId", ids["telemetry.devDeviceId"]); err != nil {
		return err
	}
	_, _ = db.Exec(`DELETE FROM ItemTable WHERE key = ?`, "cursorai/serverConfig")
	return nil
}

// WriteAuth 写入登录态五键。
func WriteAuth(p *Paths, email, token string) error {
	token = normalizeToken(token)
	email = strings.TrimSpace(email)
	if token == "" {
		return fmt.Errorf("AccessToken 为空")
	}
	db, err := sql.Open("sqlite", p.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS ItemTable (key TEXT PRIMARY KEY, value TEXT)`); err != nil {
		return err
	}
	pairs := [][2]string{
		{"cursorAuth/accessToken", token},
		{"cursorAuth/refreshToken", token},
		{"cursorAuth/cachedEmail", email},
		{"cursor.accessToken", token},
		{"cursor.email", email},
	}
	for _, kv := range pairs {
		if err := upsertItem(db, kv[0], kv[1]); err != nil {
			return err
		}
	}
	return nil
}
