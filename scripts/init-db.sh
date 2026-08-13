#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-root}"
PG_CONTAINER="${PG_CONTAINER:-ks-postgres}"
export PGPASSWORD

if command -v psql >/dev/null 2>&1; then
  run_sql() {
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$1" -v ON_ERROR_STOP=1 -f "$2"
  }
else
  run_sql() {
    docker exec -i "$PG_CONTAINER" psql -U "$PGUSER" -d "$1" -v ON_ERROR_STOP=1 < "$2"
  }
fi

echo "==> 创建数据库"
run_sql postgres "$ROOT_DIR/sql/01_init_databases.sql"

echo "==> 初始化 user 库"
run_sql shop_user "$ROOT_DIR/sql/02_user.sql"

echo "==> 初始化权限表"
run_sql shop_user "$ROOT_DIR/sql/09_permissions.sql"

echo "==> 初始化 goods 库"
run_sql shop_goods "$ROOT_DIR/sql/03_goods.sql"

echo "==> 初始化 cart 库"
run_sql shop_cart "$ROOT_DIR/sql/04_cart.sql"

echo "==> 初始化 order 库"
run_sql shop_order "$ROOT_DIR/sql/05_order.sql"

echo "==> 初始化 inventory 库"
run_sql shop_inventory "$ROOT_DIR/sql/06_inventory.sql"

echo "==> 初始化 payment 库"
run_sql shop_payment "$ROOT_DIR/sql/07_payment.sql"

echo "==> 完成"
