package cursor_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/wifi504/cursor-login-cli/cli/internal/cursor"

	_ "modernc.org/sqlite"
)

func TestWriteAuthAndReset(t *testing.T) {
	dir := t.TempDir()
	gs := filepath.Join(dir, "User", "globalStorage")
	if err := os.MkdirAll(gs, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(gs, "state.vscdb")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`CREATE TABLE ItemTable (key TEXT PRIMARY KEY, value TEXT)`); err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()

	p := &cursor.Paths{
		Root:    dir,
		Storage: filepath.Join(gs, "storage.json"),
		DB:      dbPath,
	}
	if err := cursor.ResetTelemetry(p); err != nil {
		t.Fatal(err)
	}
	if err := cursor.WriteAuth(p, "a@ex.com", "user%3A%3Atoken123"); err != nil {
		t.Fatal(err)
	}

	sqlDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var token string
	if err := sqlDB.QueryRow(`SELECT value FROM ItemTable WHERE key = 'cursorAuth/accessToken'`).Scan(&token); err != nil {
		t.Fatal(err)
	}
	if token != "token123" {
		t.Fatalf("token=%q", token)
	}
	var email string
	if err := sqlDB.QueryRow(`SELECT value FROM ItemTable WHERE key = 'cursor.email'`).Scan(&email); err != nil {
		t.Fatal(err)
	}
	if email != "a@ex.com" {
		t.Fatalf("email=%q", email)
	}
	if _, err := os.Stat(p.Storage); err != nil {
		t.Fatal(err)
	}
}
