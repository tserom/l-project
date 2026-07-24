#!/usr/bin/env bash
# Build sales-front and produce a Windows zip recipients can unzip and double-click.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

DIST_DIR="$ROOT/dist"
STAGING_PARENT="$DIST_DIR/.pack-staging-sales-front"
STAGING="$STAGING_PARENT/sales-front-windows"
ZIP="$DIST_DIR/sales-front-windows.zip"
PACK_SRC="$ROOT/apps/sales-front/pack"

echo "==> Building sales-front..."
(
  cd "$ROOT/apps/sales-front"
  CI=true pnpm install
  pnpm build
)

echo "==> Staging Windows bundle..."
rm -rf "$STAGING_PARENT"
mkdir -p "$STAGING/web"
cp -R "$ROOT/apps/sales-front/dist/." "$STAGING/web/"
cp "$PACK_SRC/start-sales.bat" "$STAGING/"
cp "$PACK_SRC/serve.ps1" "$STAGING/"
cp "$PACK_SRC/README.txt" "$STAGING/"

echo "==> Creating zip..."
mkdir -p "$DIST_DIR"
rm -f "$ZIP"
(
  cd "$STAGING_PARENT"
  if command -v zip >/dev/null 2>&1; then
    zip -r -q "$ZIP" "sales-front-windows"
  else
    # macOS fallback
    ditto -c -k --sequesterRsrc --keepParent "sales-front-windows" "$ZIP"
  fi
)

rm -rf "$STAGING_PARENT"

echo "Done: $ZIP"
echo "Send this zip to Windows users. They unzip and run start-sales.bat"
