#!/usr/bin/env bash
# 交叉编译 cursor-login CLI 全平台产物到 dist/，并同步到 admin-server/releases/bin/ 供服务端 embed。
set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p dist
export CGO_ENABLED=0

VER=dev
if tag=$(git describe --exact-match --tags HEAD 2>/dev/null); then
  VER="$tag"
fi
echo "version=${VER}"
LDFLAGS="-X main.version=${VER}"

build_one() {
  local goos="$1" goarch="$2" ext="${3:-}"
  local out="dist/cursor-login-${goos}-${goarch}${ext}"
  echo "编译 ${out}"
  GOOS="$goos" GOARCH="$goarch" go build -ldflags "$LDFLAGS" -o "$out" .
}

build_one windows amd64 .exe
build_one windows arm64 .exe
build_one darwin amd64
build_one darwin arm64
build_one linux amd64
build_one linux arm64

EMBED_DIR="$(cd .. && pwd)/admin-server/releases/bin"
mkdir -p "$EMBED_DIR"
find "$EMBED_DIR" -maxdepth 1 -type f ! -name '.gitkeep' -delete
cp -f dist/cursor-login-* "$EMBED_DIR/"
echo "CLI 全平台编译完成 -> cli/dist/ ，并已同步到 admin-server/releases/bin/"
