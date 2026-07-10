#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"

ROOT="$(pack_root)"
cd "$ROOT"

DIST_DIR="$ROOT/dist"
STAGING_PARENT="$DIST_DIR/.pack-staging-windows"
STAGING="$STAGING_PARENT/stock-inventory-windows"
ZIP="$DIST_DIR/stock-inventory-windows.zip"

is_windows_host() {
  case "$(uname -s)" in
    MINGW* | MSYS* | CYGWIN*) return 0 ;;
  esac
  [[ "${OS:-}" == Windows_NT ]]
}

echo "==> Building frontend (embed)..."
pack_build_front "$ROOT"

echo "==> Building windows/amd64 binaries..."
mkdir -p "$ROOT/bin"

build_center() {
  cd "$ROOT/apps/stock-center"
  if is_windows_host; then
    go build -o "$ROOT/bin/stock-center.exe" ./cmd/server
  else
    GOOS=windows GOARCH=amd64 go build -o "$ROOT/bin/stock-center.exe" ./cmd/server
  fi
}

build_manage() {
  cd "$ROOT/apps/stock-manage"
  if is_windows_host; then
    go build -o "$ROOT/bin/stock-manage.exe" ./cmd/server
  else
    GOOS=windows GOARCH=amd64 go build -o "$ROOT/bin/stock-manage.exe" ./cmd/server
  fi
}

build_center
build_manage

echo "==> Staging runtime bundle..."
pack_stage_dirs "$STAGING"
pack_stage_config "$STAGING" "$ROOT"
cp "$ROOT/bin/stock-center.exe" "$STAGING/bin/"
cp "$ROOT/bin/stock-manage.exe" "$STAGING/bin/"
cp "$ROOT/scripts/launcher/start.bat" "$STAGING/scripts/launcher/"
pack_write_readme "$STAGING" "Windows (amd64)" "scripts\\launcher\\start.bat"

echo "==> Creating zip..."
pack_create_zip "$STAGING" "$ZIP"

echo "Done. Unzip on Windows, configure .env, then run scripts\\launcher\\start.bat"
