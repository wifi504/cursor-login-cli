# 发版流水线：前端 embed → CLI（同步 releases/bin）→ 服务端交叉编译
chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

function Step([string]$title) {
  Write-Host ""
  Write-Host "======== $title ========" -ForegroundColor Cyan
}

Step "1/3 打包前端资源"
& powershell.exe -NoProfile -ExecutionPolicy Bypass -Command @"
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '$root\admin-web\dist'
Get-ChildItem -Path '$root\admin-server\web\dist' -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path '$root\admin-server\web\dist' | Out-Null
# 占位，避免前端构建失败时 go:embed 因空目录报错
Set-Content -Path '$root\admin-server\web\dist\index.html' -Encoding UTF8 -Value '<!DOCTYPE html><html><body><p>frontend build pending</p></body></html>'
Set-Location '$root\admin-web'
pnpm build
if (`$LASTEXITCODE -ne 0) { exit `$LASTEXITCODE }
Copy-Item -Path '$root\admin-web\dist\*' -Destination '$root\admin-server\web\dist' -Recurse -Force
Write-Host '前端已构建并复制到 admin-server/web/dist'
"@
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Step "2/3 编译全平台 CLI"
& powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$root\cli\scripts\build-all.ps1"
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Step "3/3 编译全平台服务端"
& powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$root\admin-server\scripts\build-all.ps1"
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "======== 发布构建全部完成 ========" -ForegroundColor Green
