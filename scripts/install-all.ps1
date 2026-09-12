# 安装 monorepo 全部依赖
chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")

function Step([string]$title) {
  Write-Host ""
  Write-Host "======== $title ========" -ForegroundColor Cyan
}

Step "1/3 安装前端依赖"
Set-Location "$root\admin-web"
pnpm install
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Step "2/3 安装服务端依赖"
Set-Location "$root\admin-server"
go mod download
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Step "3/3 安装 CLI 依赖"
Set-Location "$root\cli"
go mod download
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "======== 依赖安装全部完成 ========" -ForegroundColor Green
