#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DIST="$ROOT/dist"
rm -rf "$DIST"
mkdir -p "$DIST"
(cd "$ROOT/frontend" && npm ci && npm run build && tar -czf "$DIST/ghostwire-client-web.tar.gz" -C build .)
for target in linux/arm64 linux/amd64 windows/amd64; do
  os=${target%/*}
  arch=${target#*/}
  ext=""
  [ "$os" = "windows" ] && ext=".exe"
  (cd "$ROOT/backend" && CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags="-s -w" -o "$DIST/ghostwire-server-$os-$arch$ext" ./cmd/ghostwire)
done
(cd "$DIST" && tar -czf ghostwire-server-linux-arm64.tar.gz ghostwire-server-linux-arm64 && tar -czf ghostwire-server-linux-amd64.tar.gz ghostwire-server-linux-amd64 && zip -q ghostwire-server-windows-amd64.zip ghostwire-server-windows-amd64.exe)
echo "Release artifacts written to $DIST"
