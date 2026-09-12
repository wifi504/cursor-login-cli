# 清理构建/运行产物，保留依赖（node_modules、go 模块缓存）与源码占位文件。
chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

function Remove-Path([string]$rel) {
  $p = Join-Path $root $rel
  if (Test-Path $p) {
    Remove-Item -LiteralPath $p -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "已删除 $rel"
  }
}

Write-Host "======== 清理构建产物 ========"

Remove-Path "admin-web\dist"
Remove-Path "admin-server\dist"
Remove-Path "admin-server\bin"
Remove-Path "admin-server\logs"
Remove-Path "admin-server\cursor-login.db"
Remove-Path "cli\dist"
Remove-Path "cli\bin"

# web/dist：清空后写回占位，保证 go:embed 仍可用
$webDist = Join-Path $root "admin-server\web\dist"
New-Item -ItemType Directory -Force -Path $webDist | Out-Null
Get-ChildItem -LiteralPath $webDist -Force | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
$placeholder = @"
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Cursor Login Admin</title>
  </head>
  <body>
    <p>
      管理端前端尚未打包。日常开发请使用 Vite（http://localhost:5173/）；发版请先执行「构建：打包前端资源」。
    </p>
  </body>
</html>
"@
Set-Content -Path (Join-Path $webDist "index.html") -Value $placeholder.TrimStart() -Encoding UTF8
Write-Host "已还原 admin-server/web/dist/index.html"

# releases/bin：只留 .gitkeep；顺带清掉 releases 根目录误放的二进制
$relBin = Join-Path $root "admin-server\releases\bin"
New-Item -ItemType Directory -Force -Path $relBin | Out-Null
Get-ChildItem -LiteralPath $relBin -Force | Where-Object { $_.Name -ne ".gitkeep" } | Remove-Item -Recurse -Force
if (-not (Test-Path (Join-Path $relBin ".gitkeep"))) {
  Set-Content -Path (Join-Path $relBin ".gitkeep") -Value "" -Encoding UTF8
}
Get-ChildItem -LiteralPath (Join-Path $root "admin-server\releases") -File -ErrorAction SilentlyContinue |
  Where-Object { $_.Name -like "cursor-login-*" } |
  ForEach-Object { Remove-Item -LiteralPath $_.FullName -Force; Write-Host "已删除 admin-server/releases/$($_.Name)" }

Write-Host ""
Write-Host "======== 清理完成（已保留 node_modules / Go 依赖） ========" -ForegroundColor Green
