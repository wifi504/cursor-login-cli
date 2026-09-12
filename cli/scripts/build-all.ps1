# 交叉编译 cursor-login CLI 全平台产物到 dist/，并同步到 admin-server/releases/bin/ 供服务端 embed。
chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

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
  $out = "dist/cursor-login-$($t.GOOS)-$($t.GOARCH)$($t.Ext)"
  Write-Host "编译 $out"
  go build -o $out .
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

$embedDir = Join-Path (Resolve-Path ..) "admin-server\releases\bin"
New-Item -ItemType Directory -Force -Path $embedDir | Out-Null
Get-ChildItem -Path $embedDir -File | Where-Object { $_.Name -ne ".gitkeep" } | Remove-Item -Force
Copy-Item -Path "dist\cursor-login-*" -Destination $embedDir -Force
Write-Host "CLI 全平台编译完成 -> cli/dist/ ，并已同步到 admin-server/releases/bin/"