
-- =====================================================================
-- 实时库存表
-- =====================================================================
CREATE TABLE IF NOT EXISTS inventories (
    id          BIGSERIAL PRIMARY KEY,
    sku_id      BIGINT      NOT NULL UNIQUE,
    inventory   BIGINT      NOT NULL DEFAULT 0,
    locked      BIGINT      NOT NULL DEFAULT 0,
    version     BIGINT      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- 库存预占记录表
-- =====================================================================
CREATE TABLE IF NOT EXISTS inventory_locks (
    id          BIGSERIAL PRIMARY KEY,
    order_sn    VARCHAR(64) NOT NULL,
    sku_id      BIGINT      NOT NULL,
    num         INT         NOT NULL DEFAULT 0,
    status      INT         NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_inventory_locks UNIQUE (order_sn, sku_id)
);

CREATE INDEX IF NOT EXISTS idx_inventory_locks_sku ON inventory_locks (sku_id);

-- =====================================================================
-- 库存流水表
-- =====================================================================
CREATE TABLE IF NOT EXISTS inventory_flows (
    id          BIGSERIAL PRIMARY KEY,
    order_sn    VARCHAR(64) NOT NULL,
    sku_id      BIGINT      NOT NULL,
    change      BIGINT      NOT NULL DEFAULT 0,
    type        VARCHAR(32) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_flows_order ON inventory_flows (order_sn);
CREATE INDEX IF NOT EXISTS idx_inventory_flows_sku ON inventory_flows (sku_id);

-- =====================================================================
-- 消息消费幂等表
-- =====================================================================
CREATE TABLE IF NOT EXISTS consumed_event (
    event_id    VARCHAR(64) PRIMARY KEY,
    order_sn    VARCHAR(64) NOT NULL DEFAULT '',
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- Mock 库存数据（与 shop_goods.goods_skus 对应）
-- =====================================================================
INSERT INTO inventories (id, sku_id, inventory, locked, version)
VALUES
    (1, 1, 50, 0, 0),
    (2, 2, 30, 0, 0),
    (3, 3, 60, 0, 0),
    (4, 4, 40, 0, 0)
ON CONFLICT (id) DO NOTHING;

SELECT setval('inventories_id_seq', GREATEST((SELECT MAX(id) FROM inventories), 1));
SELECT setval('inventory_locks_id_seq', GREATEST((SELECT MAX(id) FROM inventory_locks), 1));
SELECT setval('inventory_flows_id_seq', GREATEST((SELECT MAX(id) FROM inventory_flows), 1));
