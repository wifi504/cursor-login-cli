#!/usr/bin/env bash
# 发版流水线：前端 embed → CLI（同步 releases/bin）→ 服务端交叉编译
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

step() { echo ""; echo "======== $1 ========"; }

step "1/3 打包前端资源"
rm -rf "$root/admin-web/dist"
mkdir -p "$root/admin-server/web/dist"
find "$root/admin-server/web/dist" -mindepth 1 -delete
# 占位，避免前端构建失败时 go:embed 因空目录报错
printf '%s\n' '<!DOCTYPE html><html><body><p>frontend build pending</p></body></html>' > "$root/admin-server/web/dist/index.html"
(cd "$root/admin-web" && pnpm build)
cp -a "$root/admin-web/dist/." "$root/admin-server/web/dist/"
echo "前端已构建并复制到 admin-server/web/dist"

step "2/3 编译全平台 CLI"
bash "$root/cli/scripts/build-all.sh"

step "3/3 编译全平台服务端"
bash "$root/admin-server/scripts/build-all.sh"

echo ""
echo "======== 发布构建全部完成 ========"
