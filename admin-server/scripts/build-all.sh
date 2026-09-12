#!/usr/bin/env bash
# 交叉编译 admin-server 全平台产物到 dist/
# 会嵌入 admin-server/releases/bin/ 下的 CLI（请先跑「构建：编译全平台 CLI」）。
set -euo pipefail
cd "$(dirname "$0")/.."

cli_count=$(find releases/bin -maxdepth 1 -type f -name 'cursor-login-*' 2>/dev/null | wc -l | tr -d ' ')
if [[ "${cli_count}" -eq 0 ]]; then
  echo "警告: releases/bin/ 下还没有 CLI 产物。发版请先编译 CLI；继续编译的服务端将无法提供 /download。" >&2
fi

mkdir -p dist
export CGO_ENABLED=0

build_one() {
  local goos="$1" goarch="$2" ext="${3:-}"
  local out="dist/admin-server-${goos}-${goarch}${ext}"
  echo "编译 ${out}"
  GOOS="$goos" GOARCH="$goarch" go build -o "$out" ./cmd/server
}

build_one windows amd64 .exe
build_one windows arm64 .exe
build_one darwin amd64
build_one darwin arm64
build_one linux amd64
build_one linux arm64

echo "服务端全平台编译完成 -> admin-server/dist/（已 embed releases/bin + web/dist）"
