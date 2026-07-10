#!/usr/bin/env bash
# Shared helpers for pack-mac.sh / pack-windows.sh

pack_root() {
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd
}

# Build frontend and copy into stock-manage embed directory.
pack_build_front() {
  local root="$1"
  make -C "$root" build-front
}

pack_stage_dirs() {
  local staging="$1"
  rm -rf "$staging"
  mkdir -p "$staging/bin"
  mkdir -p "$staging/apps/stock-center"
  mkdir -p "$staging/apps/stock-manage"
  mkdir -p "$staging/scripts/launcher"
  mkdir -p "$staging/scripts/mysql"
}

pack_stage_config() {
  local staging="$1"
  local root="$2"
  cp "$root/apps/stock-center/.env.example" "$staging/apps/stock-center/.env"
  cp "$root/apps/stock-manage/.env.example" "$staging/apps/stock-manage/.env"
  cp "$root/scripts/mysql/init-databases.sql" "$staging/scripts/mysql/"
}

pack_write_readme() {
  local staging="$1"
  local platform="$2"
  local launcher="$3"

  cat >"$staging/README.txt" <<EOF
不锈钢库存系统 — ${platform} 运行时包
================================

首次部署：
1. 安装并启动 MySQL 8
2. 建库：mysql -u root -p < scripts/mysql/init-databases.sql
3. 修改 apps/stock-center/.env 与 apps/stock-manage/.env 中的数据库密码
4. 启动：${launcher}

浏览器入口：http://localhost:8082

详细说明见仓库 docs/ 目录（开发机）或联系交付方。
EOF
}

pack_create_zip() {
  local staging="$1"
  local zip_path="$2"
  rm -f "$zip_path"
  (cd "$(dirname "$staging")" && zip -rq "$zip_path" "$(basename "$staging")")
  echo "Created: $zip_path"
}
