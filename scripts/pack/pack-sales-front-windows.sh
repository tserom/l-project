#!/usr/bin/env bash
# Build sales-front + sales-manage.exe and produce a Windows zip
# (recipient needs MySQL; unzip and double-click start-sales.bat).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

DIST_DIR="$ROOT/dist"
STAGING_PARENT="$DIST_DIR/.pack-staging-sales-front"
STAGING="$STAGING_PARENT/sales-front-windows"
ZIP="$DIST_DIR/sales-front-windows.zip"
PACK_SRC="$ROOT/apps/sales-front/pack"
MANAGE_APP="$ROOT/apps/sales-manage"

is_windows_host() {
  case "$(uname -s)" in
    MINGW* | MSYS* | CYGWIN*) return 0 ;;
  esac
  [[ "${OS:-}" == Windows_NT ]]
}

python_can_zip() {
  "$@" -c "import zipfile" >/dev/null 2>&1
}

zip_with_python() {
  local runner="$1"
  shift
  "$runner" "$@" - "$ZIP_PATH" "$FOLDER" <<'PY'
import os, sys, zipfile
dst, src = sys.argv[1], sys.argv[2]
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
  exit 1
}

echo "==> Building sales-front..."
(
  cd "$ROOT/apps/sales-front"
  CI=true pnpm install
  pnpm build
)

echo "==> Building sales-manage windows/amd64..."
mkdir -p "$ROOT/bin"
(
  cd "$MANAGE_APP"
  if is_windows_host; then
    go build -o "$ROOT/bin/sales-manage.exe" ./cmd/server
  else
    GOOS=windows GOARCH=amd64 go build -o "$ROOT/bin/sales-manage.exe" ./cmd/server
  fi
)

echo "==> Staging Windows bundle..."
rm -rf "$STAGING_PARENT"
mkdir -p "$STAGING"
cp -R "$ROOT/apps/sales-front/dist/." "$STAGING/"
cp "$PACK_SRC/start-sales.bat" "$STAGING/"
cp "$PACK_SRC/serve.ps1" "$STAGING/"
cp "$PACK_SRC/README.txt" "$STAGING/"
cp "$ROOT/bin/sales-manage.exe" "$STAGING/"
cp "$MANAGE_APP/.env.example" "$STAGING/.env"
cp "$ROOT/scripts/mysql/init-sales-manage.sql" "$STAGING/init-sales-manage.sql"

echo "==> Creating zip..."
mkdir -p "$DIST_DIR"
rm -f "$ZIP"
(
  cd "$STAGING_PARENT"
  create_sales_front_zip "$ZIP" "sales-front-windows"
)

rm -rf "$STAGING_PARENT"

echo "Done: $ZIP"
echo "Recipient: create DB (init-sales-manage.sql), edit .env, run start-sales.bat"
