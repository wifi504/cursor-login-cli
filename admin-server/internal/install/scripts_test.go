package install

import (
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	m := Commands("https://example.com/")
	iu := m["install_unix"]
	if strings.Contains(iu, "| sh") {
		t.Fatalf("install_unix must not use curl|sh pipe: %q", iu)
	}
	if !strings.Contains(iu, "mktemp") || !strings.Contains(iu, `sh "$tmp"`) {
		t.Fatalf("install_unix should download then sh temp file: %q", iu)
	}
	if !strings.HasPrefix(iu, "(") || !strings.HasSuffix(iu, ")") {
		t.Fatalf("install_unix should run in subshell: %q", iu)
	}
	if !strings.Contains(iu, "https://example.com/install.sh") || !strings.Contains(iu, "exit $e") {
		t.Fatalf("install_unix=%q", iu)
	}
	uu := m["uninstall_unix"]
	if strings.Contains(uu, "| sh") || !strings.Contains(uu, "uninstall.sh") || !strings.HasPrefix(uu, "(") {
		t.Fatalf("uninstall_unix=%q", uu)
	}
	iw := m["install_windows"]
	if strings.Contains(iw, "| iex") {
		t.Fatalf("install_windows must not use irm|iex pipe: %q", iw)
	}
	if strings.Contains(iw, "exit $LASTEXITCODE") {
		t.Fatalf("install_windows must not exit current host: %q", iw)
	}
	if !strings.Contains(iw, "Invoke-WebRequest") || !strings.Contains(iw, "OutFile") || !strings.Contains(iw, "powershell.exe") || !strings.Contains(iw, "-File $tmp") {
		t.Fatalf("install_windows should download then run temp ps1: %q", iw)
	}
	if !strings.Contains(iw, "throw") {
		t.Fatalf("install_windows should throw on failure: %q", iw)
	}
	uw := m["uninstall_windows"]
	if !strings.Contains(uw, "uninstall.ps1") || strings.Contains(uw, "exit $LASTEXITCODE") {
		t.Fatalf("uninstall_windows=%q", uw)
	}
}

func TestInstallSH_AtomicAndMarker(t *testing.T) {
	sh := InstallSH("https://api.example.com")
	if strings.Contains(sh, "正在清理旧安装") {
		t.Fatal("install must not uninstall first")
	}
	if !strings.Contains(sh, `.tmp.$$`) {
		t.Fatal("expected temp download path")
	}
	if !strings.Contains(sh, `mv "$TMP" "$DEST"`) {
		t.Fatal("expected atomic mv replace")
	}
	if !strings.Contains(sh, "下载失败，已保留原有安装") {
		t.Fatal("expected keep-old-on-failure message")
	}
	if !strings.Contains(sh, "# >>> cursor-login >>>") || !strings.Contains(sh, "# <<< cursor-login <<<") {
		t.Fatal("expected marker block")
	}
	if !strings.Contains(sh, `export PATH="${BIN_DIR}:$PATH"`) {
		t.Fatal("expected PATH export in marker")
	}
	if !strings.Contains(sh, `export CURSOR_LOGIN_API=`) {
		t.Fatal("expected CURSOR_LOGIN_API in marker")
	}
	if !strings.Contains(sh, "cursor_login_strip_rc") || !strings.Contains(sh, "cursor_login_write_rc") {
		t.Fatal("expected strip/write helpers")
	}
	if !strings.Contains(sh, ".bash_profile") {
		t.Fatal("expected bash_profile handling")
	}
	if !strings.Contains(sh, `SHELL_NAME=`) || !strings.Contains(sh, `"$OS" = "darwin"`) {
		t.Fatal("expected macOS/zsh fresh-user .zprofile bootstrap")
	}
	if !strings.Contains(sh, `touch "$HOME/.zprofile"`) {
		t.Fatal("expected create .zprofile for fresh zsh/macOS")
	}
	if !strings.Contains(sh, "重复安装会覆盖升级") {
		t.Fatal("expected overwrite-upgrade note")
	}
}

func TestUninstallSH_MarkerCleanup(t *testing.T) {
	sh := UninstallSH("https://api.example.com")
	if !strings.Contains(sh, "cursor_login_strip_rc") {
		t.Fatal("expected strip helper")
	}
	if !strings.Contains(sh, "# >>> cursor-login >>>") {
		t.Fatal("expected marker begin in strip logic")
	}
	if !strings.Contains(sh, `rm -f "${BIN_DIR}/cursor-login"`) {
		t.Fatal("expected binary removal")
	}
	if !strings.Contains(sh, `cursor-login.tmp.*`) {
		t.Fatal("expected cleanup of download temp binaries")
	}
	if !strings.Contains(sh, `.cursor-login.tmp`) || !strings.Contains(sh, `.cursor-login.tmp.2`) {
		t.Fatal("expected cleanup of rc temp files")
	}
	for _, f := range []string{".bashrc", ".bash_profile", ".zshrc", ".zprofile", ".profile"} {
		if !strings.Contains(sh, f) {
			t.Fatalf("uninstall should clean %s", f)
		}
	}
}

func TestInstallPS1_AtomicAndBroadcast(t *testing.T) {
	ps := InstallPS1("https://api.example.com")
	if strings.Contains(ps, "正在清理旧安装") {
		t.Fatal("install must not uninstall first")
	}
	if !strings.Contains(ps, `$Dest.tmp`) && !strings.Contains(ps, `"$Dest.tmp"`) {
		t.Fatal("expected temp download file")
	}
	if !strings.Contains(ps, "Move-Item") {
		t.Fatal("expected Move-Item atomic replace")
	}
	if !strings.Contains(ps, "下载失败，已保留原有安装") {
		t.Fatal("expected keep-old-on-failure message")
	}
	if strings.Contains(ps, "Add-Type") || strings.Contains(ps, "SendMessageTimeout") {
		t.Fatal("must not embed fragile Add-Type here-string")
	}
	if !strings.Contains(ps, "重启 IDE") {
		t.Fatal("expected IDE restart hint")
	}
	if !strings.Contains(ps, "Test-CursorLoginPathContains") {
		t.Fatal("expected exact PATH membership check")
	}
	if strings.Contains(ps, `$userPath -notlike "*$BinDir*"`) || strings.Contains(ps, `$env:Path -notlike "*$BinDir*"`) {
		t.Fatal("must not use substring PATH matching")
	}
}

func TestUninstallPS1_DeletesEnvKey(t *testing.T) {
	ps := UninstallPS1("https://api.example.com")
	if !strings.Contains(ps, `Remove-ItemProperty -Path "HKCU:\Environment" -Name "CURSOR_LOGIN_API"`) {
		t.Fatal("expected registry key removal")
	}
	if !strings.Contains(ps, `[NullString]::Value`) {
		t.Fatal("expected NullString env delete")
	}
	if strings.Contains(ps, "Add-Type") || strings.Contains(ps, "Publish-CursorLoginEnvChange") {
		t.Fatal("must not embed fragile Add-Type broadcast helper")
	}
}
