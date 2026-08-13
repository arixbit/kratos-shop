#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-root}"
PG_CONTAINER="${PG_CONTAINER:-ks-postgres}"
export PGPASSWORD

echo "==> 生成演示数据（100 用户 / 50 商品 / ~400 订单）"
python3 "$ROOT_DIR/scripts/generate_seed_demo.py"

if command -v psql >/dev/null 2>&1; then
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d postgres -v ON_ERROR_STOP=1 -f "$ROOT_DIR/sql/08_seed_demo.sql"
else
  docker exec -i "$PG_CONTAINER" psql -U "$PGUSER" -d postgres -v ON_ERROR_STOP=1 < "$ROOT_DIR/sql/08_seed_demo.sql"
fi

echo "==> 演示数据已写入"
