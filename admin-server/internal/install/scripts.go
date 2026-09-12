package install

import (
	"fmt"
	"strings"
)

// Commands 返回 Cursor Login Admin 可复制的一键安装/卸载命令。
func Commands(baseURL string) map[string]string {
	base := strings.TrimRight(baseURL, "/")
	return map[string]string{
		"install_unix":      fmt.Sprintf("curl -fsSL %s/install.sh | sh", base),
		"uninstall_unix":    fmt.Sprintf("curl -fsSL %s/uninstall.sh | sh", base),
		"install_windows":   fmt.Sprintf("irm %s/install.ps1 | iex", base),
		"uninstall_windows": fmt.Sprintf("irm %s/uninstall.ps1 | iex", base),
	}
}

// unixUninstallBody 卸载步骤（不含 shebang / 收尾文案），供安装前清理复用。
const unixUninstallBody = `BIN_DIR="${HOME}/.local/bin"
rm -f "${BIN_DIR}/cursor-login"
for f in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
  if [ -f "$f" ]; then
    sed -i.bak '/CURSOR_LOGIN_API=/d' "$f" 2>/dev/null || sed -i '' '/CURSOR_LOGIN_API=/d' "$f"
  fi
done
unset CURSOR_LOGIN_API || true
`

// windowsUninstallBody 卸载步骤（不含收尾文案），供安装前清理复用。
const windowsUninstallBody = `$RootDir = Join-Path $env:LOCALAPPDATA "cursor-login"
$BinDir = Join-Path $RootDir "bin"
Remove-Item -Force -ErrorAction SilentlyContinue (Join-Path $BinDir "cursor-login.exe")
Remove-Item -Force -ErrorAction SilentlyContinue (Join-Path $BinDir "cursor-login.cmd")
Remove-Item -Force -Recurse -ErrorAction SilentlyContinue $BinDir
Remove-Item -Force -Recurse -ErrorAction SilentlyContinue $RootDir

$MarkerBegin = "# >>> cursor-login >>>"
$MarkerEnd = "# <<< cursor-login <<<"
$NL = [Environment]::NewLine
foreach ($Prof in @(
  $PROFILE
  (Join-Path $HOME "Documents\PowerShell\Microsoft.PowerShell_profile.ps1")
  (Join-Path $HOME "Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1")
) | Where-Object { $_ } | Select-Object -Unique) {
  if (-not (Test-Path $Prof)) { continue }
  $Existing = Get-Content -Path $Prof -Raw -ErrorAction SilentlyContinue
  if ($Existing -and $Existing -match [regex]::Escape($MarkerBegin)) {
    $Existing = [regex]::Replace($Existing, "(?s)\r?\n?" + [regex]::Escape($MarkerBegin) + ".*?" + [regex]::Escape($MarkerEnd) + "\r?\n?", $NL)
    Set-Content -Path $Prof -Value $Existing.TrimStart() -Encoding UTF8
  }
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath) {
  $parts = $userPath.Split(';') | Where-Object { $_ -and ($_ -ne $BinDir) }
  [Environment]::SetEnvironmentVariable("Path", ($parts -join ';'), "User")
}

[Environment]::SetEnvironmentVariable("CURSOR_LOGIN_API", $null, "User")
Remove-Item Env:CURSOR_LOGIN_API -ErrorAction SilentlyContinue
if (Get-Command cursor-login -CommandType Function -ErrorAction SilentlyContinue) {
  Remove-Item -Path Function:cursor-login -ErrorAction SilentlyContinue
}
`

func InstallSH(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`#!/bin/sh
set -e
API_URL="%s"

echo "正在清理旧安装（如有）..."
%s

BIN_DIR="${HOME}/.local/bin"
mkdir -p "$BIN_DIR"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac
case "$OS" in
  linux|darwin) ;;
  *) echo "不支持的操作系统: $OS"; exit 1 ;;
esac

NAME="cursor-login-${OS}-${ARCH}"
URL="${API_URL}/download/${NAME}"
DEST="${BIN_DIR}/cursor-login"
echo "正在下载 Cursor Login CLI ..."
curl -fsSL "$URL" -o "$DEST"
chmod +x "$DEST"

ENV_LINE="export CURSOR_LOGIN_API=\"${API_URL}\""
for f in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
  if [ -f "$f" ] || [ "$f" = "$HOME/.profile" ]; then
    touch "$f"
    if grep -q 'CURSOR_LOGIN_API=' "$f" 2>/dev/null; then
      # shellcheck disable=SC2016
      sed -i.bak '/CURSOR_LOGIN_API=/d' "$f" 2>/dev/null || sed -i '' '/CURSOR_LOGIN_API=/d' "$f"
    fi
    echo "$ENV_LINE" >> "$f"
  fi
done

# 当前会话也导出，方便立即使用
export CURSOR_LOGIN_API="$API_URL"

echo "已安装 Cursor Login CLI 到 ${DEST}"
echo "请打开新终端（或 source 你的 shell 配置）后执行: cursor-login"
if ! echo ":$PATH:" | grep -q ":${BIN_DIR}:"; then
  echo "提示: 如需直接运行，请将 ${BIN_DIR} 加入 PATH"
fi
`, base, unixUninstallBody)
}

func UninstallSH(baseURL string) string {
	_ = baseURL
	return `#!/bin/sh
set -e
` + unixUninstallBody + `echo "已卸载 Cursor Login CLI"
`
}

func InstallPS1(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
$ApiUrl = "%s"

Write-Host "正在清理旧安装（如有）..."
%s

$BinDir = Join-Path $env:LOCALAPPDATA "cursor-login\bin"
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$Arch = if ([Environment]::Is64BitOperatingSystem) {
  if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else { "amd64" }
$Name = "cursor-login-windows-$Arch.exe"
$Url = "$ApiUrl/download/$Name"
$Dest = Join-Path $BinDir "cursor-login.exe"
Write-Host "正在下载 Cursor Login CLI ..."
Invoke-WebRequest -Uri $Url -OutFile $Dest

[Environment]::SetEnvironmentVariable("CURSOR_LOGIN_API", $ApiUrl, "User")
$env:CURSOR_LOGIN_API = $ApiUrl

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath) { $userPath = "" }
if ($userPath -notlike "*$BinDir*") {
  [Environment]::SetEnvironmentVariable("Path", ($userPath.TrimEnd(';') + ";" + $BinDir), "User")
  $env:Path = $env:Path + ";" + $BinDir
}

Write-Host "已安装 Cursor Login CLI 到 $Dest"
Write-Host "请新开终端（cmd / Windows Terminal / pwsh / Git Bash）后执行: cursor-login"
`, base, windowsUninstallBody)
}

func UninstallPS1(baseURL string) string {
	_ = baseURL
	return `$ErrorActionPreference = "Stop"
` + windowsUninstallBody + `
Write-Host "已卸载 Cursor Login CLI"
`
}
