#!/usr/bin/env bash
# Build sales-front and produce a Windows zip recipients can unzip and double-click.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

# Ensure standard POSIX bin dirs are present. When this script is launched from a
# Windows-native `make` (GnuWin32), the child bash may inherit a stripped PATH and
# fail to locate zip/python3. Prepending these keeps macOS and WSL/Windows behavior
# consistent.
export PATH="/usr/local/bin:/usr/bin:/bin:$PATH"

DIST_DIR="$ROOT/dist"
STAGING_PARENT="$DIST_DIR/.pack-staging-sales-front"
STAGING="$STAGING_PARENT/sales-front-windows"
ZIP="$DIST_DIR/sales-front-windows.zip"
PACK_SRC="$ROOT/apps/sales-front/pack"

# Windows App Execution Aliases (WindowsApps/python3.exe) advertise as python3 but
# exit 49 without a real interpreter. Only treat a binary as usable if zipfile imports.
python_can_zip() {
  "$@" -c "import zipfile" >/dev/null 2>&1
}

zip_with_python() {
  local runner="$1"
  shift
  "$runner" "$@" - "$ZIP_PATH" "$FOLDER" <<'PY'
import os, sys, zipfile
dst, src = sys.argv[1], sys.argv[2]
# Keep top-level folder in the archive (same as `zip -r folder`).
root_for_arc = os.path.dirname(src) or "."
with zipfile.ZipFile(dst, "w", zipfile.ZIP_DEFLATED) as z:
    for root, _, files in os.walk(src):
        for name in files:
            p = os.path.join(root, name)
            z.write(p, os.path.relpath(p, root_for_arc))
PY
}

create_sales_front_zip() {
  local zip_path="$1"
  local folder="$2"
  ZIP_PATH="$zip_path"
  FOLDER="$folder"

  if command -v zip >/dev/null 2>&1; then
    zip -r -q "$zip_path" "$folder"
    return 0
  fi

  if command -v python3 >/dev/null 2>&1 && python_can_zip python3; then
    zip_with_python python3
    return 0
  fi

  if command -v python >/dev/null 2>&1 && python_can_zip python; then
    zip_with_python python
    return 0
  fi

  # Windows py launcher (when installed); not the WindowsApps stub.
  if command -v py >/dev/null 2>&1 && python_can_zip py -3; then
    zip_with_python py -3
    return 0
  fi

  if command -v powershell.exe >/dev/null 2>&1; then
    local win_zip="$zip_path"
    if command -v cygpath >/dev/null 2>&1; then
      win_zip="$(cygpath -w "$zip_path")"
    fi
    powershell.exe -NoProfile -Command \
      "Compress-Archive -Path '$folder' -DestinationPath '$win_zip' -Force"
    return 0
  fi

  if command -v ditto >/dev/null 2>&1; then
    ditto -c -k --sequesterRsrc --keepParent "$folder" "$zip_path"
    return 0
  fi

  echo "ERROR: no zip tool found (need zip, a real python, powershell, or ditto)" >&2
  echo "Hint on Windows: install Git for Windows (powershell fallback) or real Python," >&2
  echo "or disable Settings → Apps → Advanced app settings → App execution aliases for python/python3." >&2
  exit 1
}

echo "==> Building sales-front..."
(
  cd "$ROOT/apps/sales-front"
  CI=true pnpm install
  pnpm build
)

echo "==> Staging Windows bundle..."
rm -rf "$STAGING_PARENT"
mkdir -p "$STAGING"
# Flat layout: index.html + assets next to start-sales.bat (easier for recipients)
cp -R "$ROOT/apps/sales-front/dist/." "$STAGING/"
cp "$PACK_SRC/start-sales.bat" "$STAGING/"
cp "$PACK_SRC/serve.ps1" "$STAGING/"
cp "$PACK_SRC/README.txt" "$STAGING/"

echo "==> Creating zip..."
mkdir -p "$DIST_DIR"
rm -f "$ZIP"
(
  cd "$STAGING_PARENT"
  create_sales_front_zip "$ZIP" "sales-front-windows"
)

rm -rf "$STAGING_PARENT"

echo "Done: $ZIP"
echo "Send this zip to Windows users. They unzip and run start-sales.bat"
