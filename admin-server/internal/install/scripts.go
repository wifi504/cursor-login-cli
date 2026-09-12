package install

import (
	"fmt"
	"strings"
)

// Commands 返回 Cursor Login Admin 可复制的一键安装/卸载命令。
// Unix/Windows 均先下载到临时文件再执行，避免 curl|sh / irm|iex 在下载失败时仍以 0 退出。
func Commands(baseURL string) map[string]string {
	base := strings.TrimRight(baseURL, "/")
	return map[string]string{
		"install_unix": fmt.Sprintf(
			`tmp=$(mktemp) && curl -fsSL %s/install.sh -o "$tmp" && sh "$tmp"; e=$?; rm -f "$tmp"; exit $e`, base),
		"uninstall_unix": fmt.Sprintf(
			`tmp=$(mktemp) && curl -fsSL %s/uninstall.sh -o "$tmp" && sh "$tmp"; e=$?; rm -f "$tmp"; exit $e`, base),
		"install_windows": fmt.Sprintf(
			`$tmp = Join-Path $env:TEMP ("cursor-login-install-" + [guid]::NewGuid().ToString() + ".ps1"); try { Invoke-WebRequest -Uri %s/install.ps1 -OutFile $tmp; powershell.exe -NoProfile -ExecutionPolicy Bypass -File $tmp; exit $LASTEXITCODE } finally { Remove-Item -Force -ErrorAction SilentlyContinue $tmp }`, base),
		"uninstall_windows": fmt.Sprintf(
			`$tmp = Join-Path $env:TEMP ("cursor-login-uninstall-" + [guid]::NewGuid().ToString() + ".ps1"); try { Invoke-WebRequest -Uri %s/uninstall.ps1 -OutFile $tmp; powershell.exe -NoProfile -ExecutionPolicy Bypass -File $tmp; exit $LASTEXITCODE } finally { Remove-Item -Force -ErrorAction SilentlyContinue $tmp }`, base),
	}
}

// unixStripMarkerFn 从 shell rc 中移除 cursor-login marker 块，并兼容清理旧版裸 CURSOR_LOGIN_API= 行。
const unixStripMarkerFn = `cursor_login_strip_rc() {
  f="$1"
  [ -f "$f" ] || return 0
  tmp="${f}.cursor-login.tmp"
  # 删除 marker 块（含起止行）
  awk -v b='# >>> cursor-login >>>' -v e='# <<< cursor-login <<<' '
    $0 == b {skip=1; next}
    $0 == e {skip=0; next}
    skip {next}
    {print}
  ' "$f" > "$tmp"
  # 兼容旧版：删除非 marker 的裸 CURSOR_LOGIN_API= 行
  sed '/CURSOR_LOGIN_API=/d' "$tmp" > "${tmp}.2"
  mv "${tmp}.2" "$f"
  rm -f "$tmp" "${f}.bak"
}
`

// unixUninstallBody 卸载步骤（不含 shebang / 收尾文案）。
const unixUninstallBody = `BIN_DIR="${HOME}/.local/bin"
rm -f "${BIN_DIR}/cursor-login"
# 清理下载中断残留的临时二进制
rm -f "${BIN_DIR}"/cursor-login.tmp.*
` + unixStripMarkerFn + `for f in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile"; do
  cursor_login_strip_rc "$f"
  # 清理 strip/改写中断可能留下的 rc 临时文件与旧 bak
  rm -f "${f}.cursor-login.tmp" "${f}.cursor-login.tmp.2" "${f}.bak"
done
unset CURSOR_LOGIN_API || true
`

func InstallSH(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`#!/bin/sh
set -e
API_URL="%s"
BIN_DIR="${HOME}/.local/bin"
MARKER_BEGIN="# >>> cursor-login >>>"
MARKER_END="# <<< cursor-login <<<"
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
TMP="${DEST}.tmp.$$"
echo "正在下载 Cursor Login CLI ..."
if ! curl -fsSL "$URL" -o "$TMP"; then
  rm -f "$TMP"
  echo "下载失败，已保留原有安装（如有）。"
  exit 1
fi
chmod +x "$TMP"
mv "$TMP" "$DEST"

%s
cursor_login_write_rc() {
  f="$1"
  [ -f "$f" ] || return 0
  cursor_login_strip_rc "$f"
  {
    echo "$MARKER_BEGIN"
    echo "export CURSOR_LOGIN_API=\"${API_URL}\""
    echo "export PATH=\"${BIN_DIR}:\$PATH\""
    echo "$MARKER_END"
  } >> "$f"
}

# 已存在的配置：幂等刷新 marker
for f in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile"; do
  if [ -f "$f" ]; then
    cursor_login_write_rc "$f"
  fi
done

# .profile：不存在则创建并写入（最小补齐；供 bash / sh 登录壳）
if [ ! -f "$HOME/.profile" ]; then
  touch "$HOME/.profile"
  cursor_login_write_rc "$HOME/.profile"
fi

# 已有 .bashrc 且无 .bash_profile：补最小登录壳，避免 bash 登录不读 bashrc
if [ -f "$HOME/.bashrc" ] && [ ! -f "$HOME/.bash_profile" ]; then
  printf '%%s\n' '[ -f "$HOME/.bashrc" ] && . "$HOME/.bashrc"' > "$HOME/.bash_profile"
fi

# macOS 默认 zsh（及 Linux 上 SHELL=zsh）不读 .profile；全新用户常无 .zshrc/.zprofile
SHELL_NAME=$(basename "${SHELL:-}")
if [ "$OS" = "darwin" ] || [ "$SHELL_NAME" = "zsh" ]; then
  if [ ! -f "$HOME/.zshrc" ] && [ ! -f "$HOME/.zprofile" ]; then
    touch "$HOME/.zprofile"
    cursor_login_write_rc "$HOME/.zprofile"
  fi
fi

# 当前会话立即可用
export CURSOR_LOGIN_API="$API_URL"
case ":$PATH:" in
  *":${BIN_DIR}:"*) ;;
  *) export PATH="${BIN_DIR}:$PATH" ;;
esac

echo "已安装 Cursor Login CLI 到 ${DEST}"
echo "请打开新终端（或 source 你的 shell 配置）后执行: cursor-login"
echo "重复安装会覆盖升级；卸载请使用管理端提供的卸载命令。"
`, base, unixStripMarkerFn)
}

func UninstallSH(baseURL string) string {
	_ = baseURL
	return `#!/bin/sh
set -e
` + unixUninstallBody + `echo "已卸载 Cursor Login CLI"
`
}

// windowsUninstallBody 卸载步骤（不含收尾文案）。
// 说明：User 级 SetEnvironmentVariable 会自行广播 WM_SETTINGCHANGE，无需 Add-Type。
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

# 删除用户环境变量键（勿用空字符串；$null 在部分环境下只会清空值）
Remove-ItemProperty -Path "HKCU:\Environment" -Name "CURSOR_LOGIN_API" -ErrorAction SilentlyContinue
[Environment]::SetEnvironmentVariable("CURSOR_LOGIN_API", [NullString]::Value, "User")
Remove-Item Env:CURSOR_LOGIN_API -ErrorAction SilentlyContinue
if (Get-Command cursor-login -CommandType Function -ErrorAction SilentlyContinue) {
  Remove-Item -Path Function:cursor-login -ErrorAction SilentlyContinue
}
`

func InstallPS1(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
$ApiUrl = "%s"
$BinDir = Join-Path $env:LOCALAPPDATA "cursor-login\bin"
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$Arch = if ([Environment]::Is64BitOperatingSystem) {
  if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else { "amd64" }
$Name = "cursor-login-windows-$Arch.exe"
$Url = "$ApiUrl/download/$Name"
$Dest = Join-Path $BinDir "cursor-login.exe"
$Tmp = "$Dest.tmp"
Write-Host "正在下载 Cursor Login CLI ..."
try {
  Invoke-WebRequest -Uri $Url -OutFile $Tmp
} catch {
  Remove-Item -Force -ErrorAction SilentlyContinue $Tmp
  Write-Host "下载失败，已保留原有安装（如有）。"
  throw
}
Move-Item -Force -Path $Tmp -Destination $Dest

[Environment]::SetEnvironmentVariable("CURSOR_LOGIN_API", $ApiUrl, "User")
$env:CURSOR_LOGIN_API = $ApiUrl

function Test-CursorLoginPathContains([string]$PathValue, [string]$Bin) {
  if ([string]::IsNullOrEmpty($PathValue)) { return $false }
  $binNorm = $Bin.TrimEnd('\')
  foreach ($p in $PathValue.Split(';')) {
    if ([string]::IsNullOrWhiteSpace($p)) { continue }
    if ($p.Trim().TrimEnd('\') -eq $binNorm) { return $true }
  }
  return $false
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath) { $userPath = "" }
if (-not (Test-CursorLoginPathContains $userPath $BinDir)) {
  if ($userPath.Trim().Length -eq 0) {
    [Environment]::SetEnvironmentVariable("Path", $BinDir, "User")
  } else {
    [Environment]::SetEnvironmentVariable("Path", ($userPath.TrimEnd(';') + ";" + $BinDir), "User")
  }
}
if (-not (Test-CursorLoginPathContains $env:Path $BinDir)) {
  $env:Path = $env:Path + ";" + $BinDir
}

Write-Host "已安装 Cursor Login CLI 到 $Dest"
Write-Host "请新开终端后执行: cursor-login"
Write-Host "若仍提示找不到命令：请关掉整个 Windows Terminal / 重启 IDE 后再试（已打开的终端不会自动刷新环境变量）。"
Write-Host "重复安装会覆盖升级；卸载请使用管理端提供的卸载命令。"
`, base)
}

func UninstallPS1(baseURL string) string {
	_ = baseURL
	return `$ErrorActionPreference = "Stop"
` + windowsUninstallBody + `
Write-Host "已卸载 Cursor Login CLI"
`
}
