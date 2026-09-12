package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const fileName = "cursor-login.db"

// OpenBesideBinary 在可执行文件同目录打开（或创建）SQLite。
// go run 时二进制在临时目录，改用当前工作目录，方便本地开发调试。
func OpenBesideBinary() (*sql.DB, string, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, "", err
	}
	path := filepath.Join(dir, fileName)

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", filepath.ToSlash(path))
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, "", fmt.Errorf("打开 SQLite 失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, "", fmt.Errorf("连接 SQLite 失败: %w", err)
	}
	if err := migrate(sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, "", err
	}
	return sqlDB, path, nil
}

func dataDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("解析可执行文件路径失败: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("解析可执行文件符号链接失败: %w", err)
	}
	dir := filepath.Dir(exe)
	// go run / go test 生成的临时二进制，数据落到当前工作目录
	if strings.Contains(filepath.ToSlash(dir), "/go-build") || strings.Contains(dir, `\go-build`) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("获取工作目录失败: %w", err)
		}
		return cwd, nil
	}
	return dir, nil
}

func migrate(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS admin (
  id            INTEGER PRIMARY KEY CHECK (id = 1),
  username      TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  token      TEXT PRIMARY KEY,
  username   TEXT NOT NULL,
  expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  name            TEXT NOT NULL,
  email           TEXT NOT NULL DEFAULT '',
  access_token    TEXT NOT NULL DEFAULT '',
  claim_code      TEXT NOT NULL UNIQUE,
  claim_limit     INTEGER NOT NULL DEFAULT 1,
  claimed_count   INTEGER NOT NULL DEFAULT 0,
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL,
  CHECK (claim_limit >= 0),
  CHECK (claimed_count >= 0)
);
`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}
	return nil
}
