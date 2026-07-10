#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"

ROOT="$(pack_root)"
cd "$ROOT"

DIST_DIR="$ROOT/dist"
STAGING_PARENT="$DIST_DIR/.pack-staging-mac"
STAGING="$STAGING_PARENT/stock-inventory-mac"
ZIP="$DIST_DIR/stock-inventory-mac.zip"

case "$(uname -m)" in
  arm64) GOARCH=arm64 ;;
  x86_64) GOARCH=amd64 ;;
  *)
    echo "ERROR: unsupported macOS arch: $(uname -m)" >&2
    exit 1
    ;;
esac

echo "==> Building frontend (embed)..."
pack_build_front "$ROOT"

echo "==> Building darwin/${GOARCH} binaries..."
mkdir -p "$ROOT/bin"
(
  cd "$ROOT/apps/stock-center"
  GOOS=darwin GOARCH="$GOARCH" go build -o "$ROOT/bin/stock-center" ./cmd/server
)
(
  cd "$ROOT/apps/stock-manage"
  GOOS=darwin GOARCH="$GOARCH" go build -o "$ROOT/bin/stock-manage" ./cmd/server
)

echo "==> Staging runtime bundle..."
pack_stage_dirs "$STAGING"
pack_stage_config "$STAGING" "$ROOT"
cp "$ROOT/bin/stock-center" "$STAGING/bin/"
cp "$ROOT/bin/stock-manage" "$STAGING/bin/"
cp "$ROOT/scripts/launcher/start.sh" "$STAGING/scripts/launcher/"
chmod +x "$STAGING/bin/stock-center" "$STAGING/bin/stock-manage" "$STAGING/scripts/launcher/start.sh"
pack_write_readme "$STAGING" "macOS (${GOARCH})" "./scripts/launcher/start.sh"

echo "==> Creating zip..."
pack_create_zip "$STAGING" "$ZIP"

echo "Done. Unzip on Mac (${GOARCH}), configure .env, then run scripts/launcher/start.sh"
