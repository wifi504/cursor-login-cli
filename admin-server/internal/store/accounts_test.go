package store_test

import (
	"path/filepath"
	"testing"

	"github.com/wifi504/cursor-login-cli/admin-server/internal/db"
	"github.com/wifi504/cursor-login-cli/admin-server/internal/store"
)

func TestAccountsCRUDAndOptimisticLock(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	sqlDB, path, err := db.OpenBesideBinary()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	t.Log("db:", filepath.Base(path))

	st := store.New(sqlDB)

	a1, err := st.CreateAccount(store.AccountInput{
		Name: "acc-a", Email: "a@ex.com", AccessToken: "tok-a", ClaimCode: "CODEA111111111111111111111111111",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a1.ClaimLimit != 1 || a1.ClaimedCount != 0 {
		t.Fatalf("defaults: %+v", a1)
	}

	_, err = st.CreateAccount(store.AccountInput{
		Name: "acc-b", ClaimCode: "CODEA111111111111111111111111111",
	})
	if err == nil {
		t.Fatal("expected unique claim_code failure")
	}

	_, err = st.UpdateClaimLimit(a1.ID, 999, 5)
	if err != store.ErrConflict {
		t.Fatalf("want conflict, got %v", err)
	}

	a2, err := st.UpdateClaimLimit(a1.ID, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if a2.ClaimLimit != 3 {
		t.Fatalf("limit=%d", a2.ClaimLimit)
	}

	// simulate claimed
	if _, err := sqlDB.Exec(`UPDATE accounts SET claimed_count = 2 WHERE id = ?`, a1.ID); err != nil {
		t.Fatal(err)
	}
	_, err = st.ResetClaimedCount(a1.ID, 0)
	if err != store.ErrConflict {
		t.Fatalf("want conflict on wrong from, got %v", err)
	}
	a3, err := st.ResetClaimedCount(a1.ID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if a3.ClaimedCount != 0 {
		t.Fatalf("claimed=%d", a3.ClaimedCount)
	}

	list, err := st.ListAccounts()
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}

	if err := st.DeleteAccount(a1.ID); err != nil {
		t.Fatal(err)
	}
}

func TestClaimPreviewAndRedeem(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	sqlDB, _, err := db.OpenBesideBinary()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	st := store.New(sqlDB)

	code := "CLAIMCODE11111111111111111111111"
	a, err := st.CreateAccount(store.AccountInput{
		Name: "acc", Email: "t@ex.com", AccessToken: "tok", ClaimCode: code,
	})
	if err != nil {
		t.Fatal(err)
	}
	lim := 2
	if _, err := st.UpdateClaimLimit(a.ID, 1, lim); err != nil {
		t.Fatal(err)
	}

	prev, err := st.PreviewByClaimCode(code)
	if err != nil || prev.Remaining != 2 {
		t.Fatalf("preview=%+v err=%v", prev, err)
	}

	r1, err := st.RedeemByClaimCode(code)
	if err != nil || r1.AccessToken != "tok" || r1.Remaining != 1 {
		t.Fatalf("redeem1=%+v err=%v", r1, err)
	}
	r2, err := st.RedeemByClaimCode(code)
	if err != nil || r2.Remaining != 0 {
		t.Fatalf("redeem2=%+v err=%v", r2, err)
	}
	_, err = st.RedeemByClaimCode(code)
	if err != store.ErrClaimExhausted {
		t.Fatalf("want exhausted, got %v", err)
	}

	_, err = st.PreviewByClaimCode("nope")
	if err != store.ErrNotFound {
		t.Fatalf("want not found, got %v", err)
	}

	emptyToken := "EMPTYTOKEN1111111111111111111111"
	_, err = st.CreateAccount(store.AccountInput{
		Name: "empty-token", Email: "e@ex.com", AccessToken: "", ClaimCode: emptyToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.PreviewByClaimCode(emptyToken)
	if err != store.ErrTokenNotReady {
		t.Fatalf("want not ready on preview (empty token), got %v", err)
	}
	_, err = st.RedeemByClaimCode(emptyToken)
	if err != store.ErrTokenNotReady {
		t.Fatalf("want not ready on redeem (empty token), got %v", err)
	}

	emptyEmail := "EMPTYEMAIL1111111111111111111111"
	_, err = st.CreateAccount(store.AccountInput{
		Name: "empty-email", Email: "", AccessToken: "tok", ClaimCode: emptyEmail,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.PreviewByClaimCode(emptyEmail)
	if err != store.ErrTokenNotReady {
		t.Fatalf("want not ready on preview (empty email), got %v", err)
	}
	_, err = st.RedeemByClaimCode(emptyEmail)
	if err != store.ErrTokenNotReady {
		t.Fatalf("want not ready on redeem (empty email), got %v", err)
	}
}
