package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	KeyInitialized      = "initialized"
	KeyTempEntry        = "temp_entry"
	KeyAdminEntry       = "admin_entry"
	KeyPublicBaseURL    = "public_base_url"
	KeySessionSecretTip = "session_cookie"
)

var ErrNotFound = errors.New("未找到")

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Get(key string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return v, err
}

func (s *Store) Set(key, value string) error {
	_, err := s.db.Exec(`
INSERT INTO settings(key, value) VALUES(?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
`, key, value)
	return err
}

func (s *Store) IsInitialized() (bool, error) {
	v, err := s.Get(KeyInitialized)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return v == "1", nil
}

func (s *Store) EnsureBootstrap() (tempEntry string, created bool, err error) {
	ok, err := s.IsInitialized()
	if err != nil {
		return "", false, err
	}
	if ok {
		entry, err := s.Get(KeyAdminEntry)
		if err != nil {
			return "", false, err
		}
		return entry, false, nil
	}

	existing, err := s.Get(KeyTempEntry)
	if err == nil && existing != "" {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", false, err
	}

	token, err := randomHex(16)
	if err != nil {
		return "", false, err
	}
	entry := "/" + token
	if err := s.Set(KeyTempEntry, entry); err != nil {
		return "", false, err
	}
	return entry, true, nil
}

func (s *Store) AdminEntry() (string, error) {
	ok, err := s.IsInitialized()
	if err != nil {
		return "", err
	}
	if !ok {
		return s.Get(KeyTempEntry)
	}
	return s.Get(KeyAdminEntry)
}

type BootstrapInput struct {
	Username     string
	Password     string
	AdminEntry   string
	PublicBaseURL string
}

func (s *Store) CompleteBootstrap(in BootstrapInput) error {
	ok, err := s.IsInitialized()
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("系统已初始化")
	}

	username := strings.TrimSpace(in.Username)
	password := in.Password
	entry := normalizeEntry(in.AdminEntry)
	baseURL := strings.TrimRight(strings.TrimSpace(in.PublicBaseURL), "/")

	if username == "" || password == "" {
		return fmt.Errorf("用户名和密码不能为空")
	}
	if len(password) < 6 {
		return fmt.Errorf("密码至少 6 位")
	}
	if entry == "" || entry == "/" {
		return fmt.Errorf("必须设置管理端安全入口路径")
	}
	if err := ValidateAdminEntry(entry); err != nil {
		return err
	}
	if err := ValidatePublicBaseURL(baseURL); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(`DELETE FROM admin`); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO admin(id, username, password_hash, created_at) VALUES(1, ?, ?, ?)`,
		username, string(hash), now); err != nil {
		return err
	}
	for _, kv := range [][2]string{
		{KeyAdminEntry, entry},
		{KeyPublicBaseURL, baseURL},
		{KeyInitialized, "1"},
	} {
		if _, err := tx.Exec(`
INSERT INTO settings(key, value) VALUES(?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
`, kv[0], kv[1]); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM settings WHERE key = ?`, KeyTempEntry); err != nil {
		return err
	}
	return tx.Commit()
}

type Admin struct {
	Username     string
	PasswordHash string
}

func (s *Store) GetAdmin() (*Admin, error) {
	var a Admin
	err := s.db.QueryRow(`SELECT username, password_hash FROM admin WHERE id = 1`).Scan(&a.Username, &a.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) VerifyAdmin(username, password string) (bool, error) {
	a, err := s.GetAdmin()
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if a.Username != username {
		return false, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)); err != nil {
		return false, nil
	}
	return true, nil
}

func (s *Store) CreateSession(username string, ttl time.Duration) (token string, expires time.Time, err error) {
	token, err = randomHex(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expires = time.Now().UTC().Add(ttl)
	_, err = s.db.Exec(`INSERT INTO sessions(token, username, expires_at) VALUES(?, ?, ?)`,
		token, username, expires.Format(time.RFC3339))
	return token, expires, err
}

func (s *Store) GetSession(token string) (username string, err error) {
	var expiresStr string
	err = s.db.QueryRow(`SELECT username, expires_at FROM sessions WHERE token = ?`, token).Scan(&username, &expiresStr)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	expires, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return "", err
	}
	if time.Now().UTC().After(expires) {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
		return "", ErrNotFound
	}
	return username, nil
}

func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (s *Store) DeleteAllSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions`)
	return err
}

// UpdateAdminEntry 修改管理端安全入口（需验证当前密码）。
func (s *Store) UpdateAdminEntry(entry, currentPassword string) error {
	ok, err := s.IsInitialized()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("系统尚未初始化")
	}
	admin, err := s.GetAdmin()
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("当前密码错误")
	}

	entry = normalizeEntry(entry)
	if entry == "" || entry == "/" {
		return fmt.Errorf("必须设置管理端安全入口路径")
	}
	if err := ValidateAdminEntry(entry); err != nil {
		return err
	}
	cur, err := s.Get(KeyAdminEntry)
	if err != nil {
		return err
	}
	if entry == cur {
		return nil
	}
	return s.Set(KeyAdminEntry, entry)
}

// UpdateAdminAccount 修改管理员用户名和/或密码（需验证当前密码；改密或改名后清空全部会话）。
func (s *Store) UpdateAdminAccount(username, newPassword, currentPassword string) (relogin bool, err error) {
	ok, err := s.IsInitialized()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("系统尚未初始化")
	}
	admin, err := s.GetAdmin()
	if err != nil {
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPassword)); err != nil {
		return false, fmt.Errorf("当前密码错误")
	}

	username = strings.TrimSpace(username)
	if username == "" {
		username = admin.Username
	}
	if username == "" {
		return false, fmt.Errorf("用户名不能为空")
	}

	hash := admin.PasswordHash
	changedPass := newPassword != ""
	if changedPass {
		if len(newPassword) < 6 {
			return false, fmt.Errorf("新密码至少 6 位")
		}
		b, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return false, err
		}
		hash = string(b)
	}
	if username == admin.Username && !changedPass {
		return false, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`UPDATE admin SET username = ?, password_hash = ? WHERE id = 1`, username, hash); err != nil {
		return false, err
	}
	if _, err := tx.Exec(`DELETE FROM sessions`); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) PublicBaseURL() (string, error) {
	v, err := s.Get(KeyPublicBaseURL)
	if errors.Is(err, ErrNotFound) {
		return "", nil
	}
	return v, err
}

func normalizeEntry(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimRight(p, "/")
	if p == "" {
		return "/"
	}
	return p
}

// ValidateAdminEntry 校验自定义安全入口，禁止与公开/保留路径冲突。
func ValidateAdminEntry(entry string) error {
	entry = normalizeEntry(entry)
	if entry == "" || entry == "/" {
		return fmt.Errorf("必须设置管理端安全入口路径")
	}
	reserved := []string{"/__dev__", "/api", "/install.sh", "/install.ps1", "/uninstall.sh", "/uninstall.ps1", "/download"}
	for _, r := range reserved {
		if entry == r || strings.HasPrefix(entry, r+"/") {
			return fmt.Errorf("不能使用保留路径 %s 作为安全入口", r)
		}
	}
	return nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidatePublicBaseURL 校验对外地址：仅域名（不带 path）；非本机/私网必须 https。
func ValidatePublicBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("必须填写对外访问地址")
	}
	if strings.Contains(raw, " ") {
		return fmt.Errorf("URL 无效")
	}
	u, err := parseURL(raw)
	if err != nil {
		return err
	}
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("不支持带 path 的对外地址，请仅使用域名")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("URL 不能包含查询参数或片段")
	}
	host := strings.ToLower(u.Hostname())
	local := host == "localhost" || host == "127.0.0.1" || host == "::1" || isPrivateHost(host)
	switch strings.ToLower(u.Scheme) {
	case "http":
		if !local {
			return fmt.Errorf("仅本机或局域网地址允许使用 http，其余必须 https")
		}
	case "https":
		// 允许
	default:
		return fmt.Errorf("URL 协议必须是 http 或 https")
	}
	return nil
}

func isPrivateHost(host string) bool {
	// 常见私网网段前缀判断
	if strings.HasPrefix(host, "10.") {
		return true
	}
	if strings.HasPrefix(host, "192.168.") {
		return true
	}
	if strings.HasPrefix(host, "172.") {
		parts := strings.Split(host, ".")
		if len(parts) >= 2 {
			var second int
			if _, err := fmt.Sscanf(parts[1], "%d", &second); err == nil && second >= 16 && second <= 31 {
				return true
			}
		}
	}
	return false
}
