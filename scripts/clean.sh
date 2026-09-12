#!/usr/bin/env bash
# 清理构建/运行产物，保留依赖（node_modules、go 模块缓存）与源码占位文件。
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

rm_path() {
  local rel="$1"
  if [[ -e "$root/$rel" ]]; then
    rm -rf "$root/$rel"
    echo "已删除 $rel"
  fi
}

echo "======== 清理构建产物 ========"

rm_path "admin-web/dist"
rm_path "admin-server/dist"
rm_path "admin-server/bin"
rm_path "admin-server/logs"
rm_path "admin-server/cursor-login.db"
rm_path "cli/dist"
rm_path "cli/bin"

# web/dist：清空后写回占位，保证 go:embed 仍可用
mkdir -p "$root/admin-server/web/dist"
find "$root/admin-server/web/dist" -mindepth 1 -delete
cat > "$root/admin-server/web/dist/index.html" <<'EOF'
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
EOF
echo "已还原 admin-server/web/dist/index.html"

# releases/bin：只留 .gitkeep；顺带清掉 releases 根目录误放的二进制
mkdir -p "$root/admin-server/releases/bin"
find "$root/admin-server/releases/bin" -mindepth 1 ! -name '.gitkeep' -delete
touch "$root/admin-server/releases/bin/.gitkeep"
find "$root/admin-server/releases" -maxdepth 1 -type f -name 'cursor-login-*' -print -delete

echo ""
echo "======== 清理完成（已保留 node_modules / Go 依赖） ========"
