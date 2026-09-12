# 交叉编译 admin-server 全平台产物到 dist/
# 会嵌入 admin-server/releases/bin/ 下的 CLI（请先跑「构建：编译全平台 CLI」）。
chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

$binDir = Join-Path (Get-Location) "releases\bin"
$cliCount = @(Get-ChildItem -Path $binDir -File -ErrorAction SilentlyContinue | Where-Object { $_.Name -like "cursor-login-*" }).Count
if ($cliCount -eq 0) {
  Write-Warning "releases/bin/ 下还没有 CLI 产物。发版请先编译 CLI；继续编译的服务端将无法提供 /download。"
}

New-Item -ItemType Directory -Force -Path dist | Out-Null
$env:CGO_ENABLED = "0"

$targets = @(
  @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
  @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" },
  @{ GOOS = "darwin"; GOARCH = "amd64"; Ext = "" },
  @{ GOOS = "darwin"; GOARCH = "arm64"; Ext = "" },
  @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
  @{ GOOS = "linux"; GOARCH = "arm64"; Ext = "" }
)

foreach ($t in $targets) {
  $env:GOOS = $t.GOOS
  $env:GOARCH = $t.GOARCH
  $out = "dist/admin-server-$($t.GOOS)-$($t.GOARCH)$($t.Ext)"
  Write-Host "编译 $out"
  go build -o $out ./cmd/server
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Write-Host "服务端全平台编译完成 -> admin-server/dist/（已 embed releases/bin + web/dist）"