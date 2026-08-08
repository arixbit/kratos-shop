#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="$ROOT_DIR/sql/migrations"

ACTION="${1:-up}"
DB_NAME="${2:-}"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-root}"

DBS=(shop_user shop_goods shop_cart shop_order shop_inventory shop_payment)
if [ -n "$DB_NAME" ]; then
  DBS=("$DB_NAME")
fi

run_migrate() {
  local db="$1"
  local url="postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$db?sslmode=disable"

  if command -v migrate >/dev/null 2>&1; then
    migrate -path "$MIGRATIONS_DIR/$db" -database "$url" "$ACTION"
    return
  fi

  if command -v docker >/dev/null 2>&1; then
    local dburl="$url"
    local docker_args=(--rm -v "$MIGRATIONS_DIR:/migrations")
    case "$(uname -s)" in
      Darwin|MINGW*|MSYS*)
        dburl="postgres://$PGUSER:$PGPASSWORD@host.docker.internal:$PGPORT/$db?sslmode=disable"
        ;;
      Linux)
        docker_args+=(--network host)
        ;;
    esac
    docker run "${docker_args[@]}" migrate/migrate \
      -path "/migrations/$db" -database "$dburl" "$ACTION"
    return
  fi

  echo "未找到 migrate CLI 或 Docker，请先安装 golang-migrate：https://github.com/golang-migrate/migrate" >&2
  exit 1
}

for db in "${DBS[@]}"; do
  echo "==> migrate $ACTION $db"
  run_migrate "$db"
done
