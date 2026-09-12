package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrConflict 表示乐观锁冲突（原始值与库中不一致）。
var ErrConflict = errors.New("数据已变更，请刷新页面")

// Account 池中的 Cursor 账号（管理端明文读写）。
type Account struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	ClaimCode    string `json:"claim_code"`
	ClaimLimit   int    `json:"claim_limit"`
	ClaimedCount int    `json:"claimed_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type AccountInput struct {
	Name        string
	Email       string
	AccessToken string
	ClaimCode   string
	ClaimLimit  *int // nil 表示创建时用默认 1；更新时 nil 表示不改配额（配额走专用接口）
}

func (s *Store) ListAccounts() ([]Account, error) {
	rows, err := s.db.Query(`
SELECT id, name, email, access_token, claim_code, claim_limit, claimed_count, created_at, updated_at
FROM accounts ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.AccessToken, &a.ClaimCode,
			&a.ClaimLimit, &a.ClaimedCount, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) GetAccount(id int64) (*Account, error) {
	var a Account
	err := s.db.QueryRow(`
SELECT id, name, email, access_token, claim_code, claim_limit, claimed_count, created_at, updated_at
FROM accounts WHERE id = ?`, id).Scan(
		&a.ID, &a.Name, &a.Email, &a.AccessToken, &a.ClaimCode,
		&a.ClaimLimit, &a.ClaimedCount, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) CreateAccount(in AccountInput) (*Account, error) {
	name := strings.TrimSpace(in.Name)
	email := strings.TrimSpace(in.Email)
	code := strings.TrimSpace(in.ClaimCode)
	token := strings.TrimSpace(in.AccessToken)
	if name == "" {
		return nil, fmt.Errorf("展示名不能为空")
	}
	if code == "" {
		return nil, fmt.Errorf("上号码不能为空")
	}
	limit := 1
	if in.ClaimLimit != nil {
		limit = *in.ClaimLimit
	}
	if limit < 0 {
		return nil, fmt.Errorf("领取配额不能为负数")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`
INSERT INTO accounts(name, email, access_token, claim_code, claim_limit, claimed_count, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, 0, ?, ?)`,
		name, email, token, code, limit, now, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("上号码已存在，请换一个")
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetAccount(id)
}

func (s *Store) UpdateAccount(id int64, in AccountInput) (*Account, error) {
	cur, err := s.GetAccount(id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	email := strings.TrimSpace(in.Email)
	code := strings.TrimSpace(in.ClaimCode)
	token := strings.TrimSpace(in.AccessToken)
	if name == "" {
		return nil, fmt.Errorf("展示名不能为空")
	}
	if code == "" {
		return nil, fmt.Errorf("上号码不能为空")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`
UPDATE accounts SET name = ?, email = ?, access_token = ?, claim_code = ?, updated_at = ?
WHERE id = ?`, name, email, token, code, now, id)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("上号码已存在，请换一个")
		}
		return nil, err
	}
	_ = cur
	return s.GetAccount(id)
}

func (s *Store) UpdateClaimLimit(id int64, from, to int) (*Account, error) {
	if to < 0 {
		return nil, fmt.Errorf("领取配额不能为负数")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`
UPDATE accounts SET claim_limit = ?, updated_at = ?
WHERE id = ? AND claim_limit = ?`, to, now, id, from)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		if _, err := s.GetAccount(id); errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	return s.GetAccount(id)
}

func (s *Store) ResetClaimedCount(id int64, from int) (*Account, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`
UPDATE accounts SET claimed_count = 0, updated_at = ?
WHERE id = ? AND claimed_count = ?`, now, id, from)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		if _, err := s.GetAccount(id); errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	return s.GetAccount(id)
}

func (s *Store) DeleteAccount(id int64) error {
	res, err := s.db.Exec(`DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
