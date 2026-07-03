#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

CENTER_PID=""
MANAGE_PID=""

cleanup() {
  if [[ -n "${MANAGE_PID}" ]]; then
    kill "${MANAGE_PID}" 2>/dev/null || true
  fi
  if [[ -n "${CENTER_PID}" ]]; then
    kill "${CENTER_PID}" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

wait_for_health() {
  local url=$1
  local name=$2
  local max=${3:-60}
  local i=0

  until curl -sf "${url}" >/dev/null; do
    sleep 1
    i=$((i + 1))
    if (( i >= max )); then
      echo "ERROR: ${name} did not become healthy at ${url}" >&2
      echo "Check MySQL is running and .env files are configured." >&2
      exit 1
    fi
  done
  echo "${name} is healthy"
}

load_env() {
  local env_file=$1
  if [[ -f "${env_file}" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "${env_file}"
    set +a
  fi
}

start_center() {
  if [[ -x "${ROOT}/bin/stock-center" ]]; then
    echo "Starting bin/stock-center..."
    (
      load_env "${ROOT}/apps/stock-center/.env"
      exec "${ROOT}/bin/stock-center"
    ) &
  else
    echo "bin/stock-center not found; run 'make build-center' or using go run..."
    (
      cd "${ROOT}/apps/stock-center"
      load_env .env
      exec go run ./cmd/server
    ) &
  fi
  CENTER_PID=$!
}

start_manage() {
  if [[ -x "${ROOT}/bin/stock-manage" ]]; then
    echo "Starting bin/stock-manage..."
    (
      load_env "${ROOT}/apps/stock-manage/.env"
      exec "${ROOT}/bin/stock-manage"
    ) &
  else
    echo "bin/stock-manage not found; run 'make build-manage' or using go run..."
    (
      cd "${ROOT}/apps/stock-manage"
      load_env .env
      exec go run ./cmd/server
    ) &
  fi
  MANAGE_PID=$!
}

start_center
wait_for_health "http://localhost:8081/health" "stock-center"

start_manage
wait_for_health "http://localhost:8082/health" "stock-manage"

echo "Opening http://localhost:8082 ..."
if command -v open >/dev/null 2>&1; then
  open "http://localhost:8082"
elif command -v xdg-open >/dev/null 2>&1; then
  xdg-open "http://localhost:8082"
fi

echo "Services running (center PID ${CENTER_PID}, manage PID ${MANAGE_PID}). Press Ctrl+C to stop."
wait
