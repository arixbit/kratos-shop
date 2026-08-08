#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "用法: $0 <备份目录>"
  echo "示例: $0 backups/20260808-120000"
  exit 1
fi

BACKUP_DIR="$1"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-root}"
PG_CONTAINER="${PG_CONTAINER:-ks-postgres}"

DBS=(shop_user shop_goods shop_cart shop_order shop_inventory shop_payment)

for db in "${DBS[@]}"; do
  if [ ! -f "$BACKUP_DIR/$db.dump" ]; then
    echo "缺少备份文件: $BACKUP_DIR/$db.dump" >&2
    exit 1
  fi
done

read -r -p "恢复将覆盖以下 6 个数据库，输入 YES 继续: " confirm
if [ "$confirm" != "YES" ]; then
  echo "已取消"
  exit 1
fi

restore_db() {
  local db="$1"
  if command -v pg_restore >/dev/null 2>&1; then
    PGPASSWORD="$PGPASSWORD" pg_restore --clean --if-exists \
      -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$db" \
      "$BACKUP_DIR/$db.dump"
  elif command -v docker >/dev/null 2>&1; then
    docker exec -i -e PGPASSWORD="$PGPASSWORD" "$PG_CONTAINER" \
      pg_restore --clean --if-exists -U "$PGUSER" -d "$db" \
      < "$BACKUP_DIR/$db.dump"
  else
    echo "需要 pg_restore 或 Docker" >&2
    exit 1
  fi
  echo "restore $db <- $BACKUP_DIR/$db.dump"
}

for db in "${DBS[@]}"; do
  restore_db "$db"
done

echo "==> 恢复完成"
