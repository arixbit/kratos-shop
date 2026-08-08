#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_ROOT="${BACKUP_ROOT:-$ROOT_DIR/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-root}"
PG_CONTAINER="${PG_CONTAINER:-ks-postgres}"

DBS=(shop_user shop_goods shop_cart shop_order shop_inventory shop_payment)
TS="$(date +%Y%m%d-%H%M%S)"
BACKUP_DIR="$BACKUP_ROOT/$TS"
mkdir -p "$BACKUP_DIR"

dump_db() {
  local db="$1"
  if command -v pg_dump >/dev/null 2>&1; then
    PGPASSWORD="$PGPASSWORD" pg_dump -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$db" -Fc \
      -f "$BACKUP_DIR/$db.dump"
  elif command -v docker >/dev/null 2>&1; then
    docker exec -e PGPASSWORD="$PGPASSWORD" "$PG_CONTAINER" \
      pg_dump -U "$PGUSER" -d "$db" -Fc > "$BACKUP_DIR/$db.dump"
  else
    echo "需要 pg_dump 或 Docker" >&2
    exit 1
  fi
  echo "backup $db -> $BACKUP_DIR/$db.dump"
}

for db in "${DBS[@]}"; do
  dump_db "$db"
done

if [ -d "$BACKUP_ROOT" ]; then
  find "$BACKUP_ROOT" -maxdepth 1 -type d -mtime +"$RETENTION_DAYS" \
    -exec rm -rf {} + 2>/dev/null || true
fi

echo "==> 备份完成: $BACKUP_DIR"
