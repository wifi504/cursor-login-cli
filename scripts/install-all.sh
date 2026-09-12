#!/usr/bin/env bash
# 安装 monorepo 全部依赖
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"

step() { echo ""; echo "======== $1 ========"; }

step "1/3 安装前端依赖"
(cd "$root/admin-web" && pnpm install)

step "2/3 安装服务端依赖"
(cd "$root/admin-server" && go mod download)

step "3/3 安装 CLI 依赖"
(cd "$root/cli" && go mod download)

echo ""
echo "======== 依赖安装全部完成 ========"
