\connect shop_goods

-- =====================================================================
-- 商品品牌表
-- =====================================================================
CREATE TABLE IF NOT EXISTS brands (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL,
    logo        VARCHAR(200) NOT NULL DEFAULT '',
    "desc"      VARCHAR(500) NOT NULL DEFAULT '',
    is_tab      BOOLEAN      NOT NULL DEFAULT FALSE,
    sort        INT          NOT NULL DEFAULT 99,
    add_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- =====================================================================
-- 商品分类表
-- =====================================================================
CREATE TABLE IF NOT EXISTS categories (
    id                 BIGSERIAL PRIMARY KEY,
    name               VARCHAR(50) NOT NULL,
    parent_category_id BIGINT      NOT NULL DEFAULT 0,
    level              INT         NOT NULL DEFAULT 1,
    is_tab             BOOLEAN     NOT NULL DEFAULT FALSE,
    sort               INT         NOT NULL DEFAULT 99,
    add_time           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories (parent_category_id);

-- =====================================================================
-- 分类-品牌关联表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_category_brands (
    id          BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    brands_id   BIGINT NOT NULL,
    add_time    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uk_goods_category_brands UNIQUE (category_id, brands_id)
);

-- =====================================================================
-- 商品类型表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_types (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
    type_code   VARCHAR(50) NOT NULL,
    name_alias  VARCHAR(50) NOT NULL DEFAULT '',
    is_virtual  BOOLEAN     NOT NULL DEFAULT FALSE,
    "desc"      VARCHAR(50) NOT NULL DEFAULT '',
    sort        INT         NOT NULL DEFAULT 99,
    add_time    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- =====================================================================
-- 商品类型-品牌关联表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_type_brands (
    id       BIGSERIAL PRIMARY KEY,
    brand_id BIGINT NOT NULL,
    type_id  BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_goods_type_brands_brand ON goods_type_brands (brand_id);
CREATE INDEX IF NOT EXISTS idx_goods_type_brands_type ON goods_type_brands (type_id);

-- =====================================================================
-- 规格参数表
-- =====================================================================
CREATE TABLE IF NOT EXISTS specifications_attrs (
    id          BIGSERIAL PRIMARY KEY,
    type_id     BIGINT       NOT NULL,
    name        VARCHAR(250) NOT NULL,
    sort        INT          NOT NULL DEFAULT 99,
    status      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_sku      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_select   BOOLEAN      NOT NULL DEFAULT FALSE,
    add_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_specifications_attrs_type ON specifications_attrs (type_id);

-- =====================================================================
-- 规格参数值表
-- =====================================================================
CREATE TABLE IF NOT EXISTS specifications_attr_values (
    id          BIGSERIAL PRIMARY KEY,
    attr_id     BIGINT       NOT NULL,
    value       VARCHAR(250) NOT NULL,
    sort        INT          NOT NULL DEFAULT 99,
    add_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_specifications_attr_values_attr ON specifications_attr_values (attr_id);

-- =====================================================================
-- 商品属性分组表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_attr_groups (
    id            BIGSERIAL PRIMARY KEY,
    goods_type_id BIGINT       NOT NULL,
    title         VARCHAR(100) NOT NULL,
    "desc"        VARCHAR(200) NOT NULL DEFAULT '',
    status        BOOLEAN      NOT NULL DEFAULT FALSE,
    sort          INT          NOT NULL DEFAULT 0,
    add_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_attr_groups_type ON goods_attr_groups (goods_type_id);

-- =====================================================================
-- 商品属性表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_attrs (
    id            BIGSERIAL PRIMARY KEY,
    goods_type_id BIGINT       NOT NULL,
    group_id      BIGINT       NOT NULL,
    title         VARCHAR(100) NOT NULL,
    "desc"        VARCHAR(200) NOT NULL DEFAULT '',
    status        BOOLEAN      NOT NULL DEFAULT FALSE,
    sort          INT          NOT NULL DEFAULT 0,
    add_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_attrs_type ON goods_attrs (goods_type_id);
CREATE INDEX IF NOT EXISTS idx_goods_attrs_group ON goods_attrs (group_id);

-- =====================================================================
-- 商品属性值表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_attr_values (
    id          BIGSERIAL PRIMARY KEY,
    attr_id     BIGINT       NOT NULL,
    group_id    BIGINT       NOT NULL,
    value       VARCHAR(100) NOT NULL,
    add_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_attr_values_attr ON goods_attr_values (attr_id);
CREATE INDEX IF NOT EXISTS idx_goods_attr_values_group ON goods_attr_values (group_id);

-- =====================================================================
-- 商品表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods (
    id               BIGSERIAL PRIMARY KEY,
    category_id      BIGINT       NOT NULL,
    brands_id        BIGINT       NOT NULL,
    type_id          BIGINT       NOT NULL,
    name             VARCHAR(100) NOT NULL,
    name_alias       VARCHAR(100) NOT NULL DEFAULT '',
    goods_sn         VARCHAR(100) NOT NULL,
    goods_tags       VARCHAR(100) NOT NULL DEFAULT '',
    market_price     BIGINT       NOT NULL DEFAULT 0,
    goods_brief      VARCHAR(100) NOT NULL DEFAULT '',
    goods_front_image VARCHAR(200) NOT NULL DEFAULT '',
    goods_images     TEXT         NOT NULL DEFAULT '[]',
    on_sale          BOOLEAN      NOT NULL DEFAULT FALSE,
    ship_free        BOOLEAN      NOT NULL DEFAULT FALSE,
    ship_id          BIGINT       NOT NULL DEFAULT 0,
    is_new           BOOLEAN      NOT NULL DEFAULT FALSE,
    is_hot           BOOLEAN      NOT NULL DEFAULT FALSE,
    click_num        BIGINT       NOT NULL DEFAULT 0,
    sold_num         BIGINT       NOT NULL DEFAULT 0,
    fav_num          BIGINT       NOT NULL DEFAULT 0,
    add_time         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_category ON goods (category_id);
CREATE INDEX IF NOT EXISTS idx_goods_brand ON goods (brands_id);
CREATE INDEX IF NOT EXISTS idx_goods_type ON goods (type_id);

-- =====================================================================
-- 商品 SKU 表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_skus (
    id               BIGSERIAL PRIMARY KEY,
    goods_id         BIGINT       NOT NULL,
    goods_sn         VARCHAR(100) NOT NULL DEFAULT '',
    goods_name       VARCHAR(100) NOT NULL DEFAULT '',
    sku_name         VARCHAR(100) NOT NULL DEFAULT '',
    sku_code         VARCHAR(100) NOT NULL DEFAULT '',
    bar_code         VARCHAR(100) NOT NULL DEFAULT '',
    price            BIGINT       NOT NULL DEFAULT 0,
    promotion_price  BIGINT       NOT NULL DEFAULT 0,
    points           BIGINT       NOT NULL DEFAULT 0,
    remarks_info     VARCHAR(100) NOT NULL DEFAULT '',
    pic              VARCHAR(500) NOT NULL DEFAULT '',
    on_sale          BOOLEAN      NOT NULL DEFAULT FALSE,
    attr_info        VARCHAR(2000) NOT NULL DEFAULT '',
    inventory        BIGINT       NOT NULL DEFAULT 0,
    add_time         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_skus_goods ON goods_skus (goods_id);

-- =====================================================================
-- 商品 SKU-规格关联表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_specification_skus (
    id               BIGSERIAL PRIMARY KEY,
    sku_id           BIGINT       NOT NULL,
    sku_code         VARCHAR(100) NOT NULL DEFAULT '',
    specification_id BIGINT       NOT NULL,
    value_id         BIGINT       NOT NULL,
    add_time         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_specification_skus_sku ON goods_specification_skus (sku_id);
CREATE INDEX IF NOT EXISTS idx_goods_specification_skus_spec ON goods_specification_skus (specification_id);
CREATE INDEX IF NOT EXISTS idx_goods_specification_skus_value ON goods_specification_skus (value_id);

-- =====================================================================
-- 商品库存表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_inventories (
    id          BIGSERIAL PRIMARY KEY,
    sku_id      BIGINT NOT NULL,
    inventory   BIGINT NOT NULL DEFAULT 0,
    add_time    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_inventories_sku ON goods_inventories (sku_id);

-- =====================================================================
-- 商品图片表
-- =====================================================================
CREATE TABLE IF NOT EXISTS goods_images (
    id          BIGSERIAL PRIMARY KEY,
    goods_id    BIGINT       NOT NULL,
    link        VARCHAR(200) NOT NULL DEFAULT '',
    position    SMALLINT     NOT NULL DEFAULT 0,
    is_master   BOOLEAN      NOT NULL DEFAULT FALSE,
    add_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_goods_images_goods ON goods_images (goods_id);

-- =====================================================================
-- 消息消费幂等表
-- =====================================================================
CREATE TABLE IF NOT EXISTS consumed_event (
    event_id    VARCHAR(64) PRIMARY KEY,
    order_sn    VARCHAR(64) NOT NULL DEFAULT '',
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- Mock 基础数据
-- =====================================================================
INSERT INTO brands (id, name, logo, "desc", is_tab, sort)
VALUES
    (1, 'Apple',   'https://example.com/apple.png',   '苹果',   TRUE, 1),
    (2, 'Huawei',  'https://example.com/huawei.png',  '华为',   TRUE, 2),
    (3, 'Xiaomi',  'https://example.com/xiaomi.png',  '小米',   TRUE, 3),
    (4, 'Lenovo',  'https://example.com/lenovo.png',  '联想',   FALSE, 4)
ON CONFLICT (id) DO NOTHING;

INSERT INTO categories (id, name, parent_category_id, level, is_tab, sort)
VALUES
    (1,  '手机数码', 0, 1, TRUE,  1),
    (2,  '电脑办公', 0, 1, TRUE,  2),
    (11, '智能手机', 1, 2, TRUE,  1),
    (12, '手机配件', 1, 2, FALSE, 2),
    (21, '笔记本电脑', 2, 2, TRUE, 1),
    (111, '全面屏手机', 11, 3, TRUE, 1),
    (211, '轻薄本', 21, 3, TRUE, 1)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_category_brands (id, category_id, brands_id)
VALUES (1, 11, 1), (2, 11, 2), (3, 11, 3), (4, 21, 4)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_types (id, name, type_code, name_alias, is_virtual, "desc", sort)
VALUES
    (1, '手机', 'PHONE', '智能手机', FALSE, '手机商品类型', 1),
    (2, '电脑', 'PC',    '笔记本电脑', FALSE, '电脑商品类型', 2)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_type_brands (id, brand_id, type_id)
VALUES (1, 1, 1), (2, 2, 1), (3, 3, 1), (4, 4, 2)
ON CONFLICT (id) DO NOTHING;

INSERT INTO specifications_attrs (id, type_id, name, sort, status, is_sku, is_select)
VALUES
    (1, 1, '颜色', 1, TRUE, TRUE, TRUE),
    (2, 1, '内存', 2, TRUE, TRUE, TRUE),
    (3, 2, '颜色', 1, TRUE, TRUE, TRUE),
    (4, 2, '硬盘', 2, TRUE, TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO specifications_attr_values (id, attr_id, value, sort)
VALUES
    (1, 1, '黑色', 1),
    (2, 1, '白色', 2),
    (3, 1, '蓝色', 3),
    (4, 2, '128G', 1),
    (5, 2, '256G', 2),
    (6, 3, '灰色', 1),
    (7, 4, '512G SSD', 1),
    (8, 4, '1T SSD', 2)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_attr_groups (id, goods_type_id, title, "desc", status, sort)
VALUES
    (1, 1, '主体', '手机主体信息', TRUE, 1),
    (2, 1, '屏幕', '手机屏幕信息', TRUE, 2),
    (3, 2, '主体', '电脑主体信息', TRUE, 1)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_attrs (id, goods_type_id, group_id, title, "desc", status, sort)
VALUES
    (1, 1, 1, '品牌', '手机品牌', TRUE, 1),
    (2, 1, 1, '型号', '手机型号', TRUE, 2),
    (3, 1, 2, '屏幕尺寸', '手机屏幕尺寸', TRUE, 1),
    (4, 2, 3, '品牌', '电脑品牌', TRUE, 1)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_attr_values (id, attr_id, group_id, value)
VALUES
    (1, 1, 1, 'Apple'),
    (2, 1, 1, 'Huawei'),
    (3, 2, 1, 'iPhone 15 Pro'),
    (4, 2, 1, 'Mate 60 Pro'),
    (5, 3, 2, '6.1 英寸'),
    (6, 4, 3, 'Lenovo')
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods (id, category_id, brands_id, type_id, name, name_alias, goods_sn, goods_tags,
                   market_price, goods_brief, goods_front_image, goods_images, on_sale, ship_free,
                   ship_id, is_new, is_hot, click_num, sold_num, fav_num)
VALUES
    (1, 111, 1, 1, 'iPhone 15 Pro', '苹果 iPhone 15 Pro', 'SN-IPHONE-001', '旗舰,新品',
     899900, 'A17 Pro 芯片', 'https://example.com/iphone15pro.jpg', '["https://example.com/iphone15pro-1.jpg","https://example.com/iphone15pro-2.jpg"]',
     TRUE, TRUE, 1, TRUE, TRUE, 1200, 300, 88),
    (2, 111, 2, 1, '华为 Mate 60 Pro', 'Huawei Mate 60 Pro', 'SN-HW-001', '旗舰,新品',
     699900, '卫星通话', 'https://example.com/mate60pro.jpg', '["https://example.com/mate60pro-1.jpg"]',
     TRUE, TRUE, 1, TRUE, FALSE, 800, 150, 45),
    (3, 211, 4, 2, '联想小新 Pro 16', 'Lenovo 小新 Pro 16', 'SN-LENOVO-001', '轻薄本',
     549900, '高性能轻薄本', 'https://example.com/xiaoxinpro16.jpg', '["https://example.com/xiaoxinpro16-1.jpg"]',
     TRUE, TRUE, 2, FALSE, TRUE, 500, 80, 20)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_skus (id, goods_id, goods_sn, goods_name, sku_name, sku_code, bar_code,
                        price, promotion_price, points, remarks_info, pic, on_sale, attr_info, inventory)
VALUES
    (1, 1, 'SN-IPHONE-001', 'iPhone 15 Pro', '黑色 128G', 'SKU-IP-001', '690000000001',
     899900, 849900, 100, '', 'https://example.com/iphone15pro-black.jpg', TRUE, '{"颜色":"黑色","内存":"128G"}', 50),
    (2, 1, 'SN-IPHONE-001', 'iPhone 15 Pro', '白色 256G', 'SKU-IP-002', '690000000002',
     999900, 949900, 120, '', 'https://example.com/iphone15pro-white.jpg', TRUE, '{"颜色":"白色","内存":"256G"}', 30),
    (3, 2, 'SN-HW-001', '华为 Mate 60 Pro', '黑色 256G', 'SKU-HW-001', '690000000003',
     699900, 679900, 90, '', 'https://example.com/mate60pro-black.jpg', TRUE, '{"颜色":"黑色","内存":"256G"}', 60),
    (4, 3, 'SN-LENOVO-001', '联想小新 Pro 16', '灰色 512G SSD', 'SKU-LENOVO-001', '690000000004',
     549900, 529900, 80, '', 'https://example.com/xiaoxinpro16-gray.jpg', TRUE, '{"颜色":"灰色","硬盘":"512G SSD"}', 40)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_specification_skus (id, sku_id, sku_code, specification_id, value_id)
VALUES
    (1, 1, 'SKU-IP-001', 1, 1),
    (2, 1, 'SKU-IP-001', 2, 4),
    (3, 2, 'SKU-IP-002', 1, 2),
    (4, 2, 'SKU-IP-002', 2, 5),
    (5, 3, 'SKU-HW-001', 1, 1),
    (6, 3, 'SKU-HW-001', 2, 5),
    (7, 4, 'SKU-LENOVO-001', 3, 6),
    (8, 4, 'SKU-LENOVO-001', 4, 7)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_inventories (id, sku_id, inventory)
VALUES (1, 1, 50), (2, 2, 30), (3, 3, 60), (4, 4, 40)
ON CONFLICT (id) DO NOTHING;

INSERT INTO goods_images (id, goods_id, link, position, is_master)
VALUES
    (1, 1, 'https://example.com/iphone15pro-1.jpg', 1, TRUE),
    (2, 1, 'https://example.com/iphone15pro-2.jpg', 2, FALSE),
    (3, 2, 'https://example.com/mate60pro-1.jpg', 1, TRUE),
    (4, 3, 'https://example.com/xiaoxinpro16-1.jpg', 1, TRUE)
ON CONFLICT (id) DO NOTHING;

SELECT setval('brands_id_seq', GREATEST((SELECT MAX(id) FROM brands), 1));
SELECT setval('categories_id_seq', GREATEST((SELECT MAX(id) FROM categories), 1));
SELECT setval('goods_category_brands_id_seq', GREATEST((SELECT MAX(id) FROM goods_category_brands), 1));
SELECT setval('goods_types_id_seq', GREATEST((SELECT MAX(id) FROM goods_types), 1));
SELECT setval('goods_type_brands_id_seq', GREATEST((SELECT MAX(id) FROM goods_type_brands), 1));
SELECT setval('specifications_attrs_id_seq', GREATEST((SELECT MAX(id) FROM specifications_attrs), 1));
SELECT setval('specifications_attr_values_id_seq', GREATEST((SELECT MAX(id) FROM specifications_attr_values), 1));
SELECT setval('goods_attr_groups_id_seq', GREATEST((SELECT MAX(id) FROM goods_attr_groups), 1));
SELECT setval('goods_attrs_id_seq', GREATEST((SELECT MAX(id) FROM goods_attrs), 1));
SELECT setval('goods_attr_values_id_seq', GREATEST((SELECT MAX(id) FROM goods_attr_values), 1));
SELECT setval('goods_id_seq', GREATEST((SELECT MAX(id) FROM goods), 1));
SELECT setval('goods_skus_id_seq', GREATEST((SELECT MAX(id) FROM goods_skus), 1));
SELECT setval('goods_specification_skus_id_seq', GREATEST((SELECT MAX(id) FROM goods_specification_skus), 1));
SELECT setval('goods_inventories_id_seq', GREATEST((SELECT MAX(id) FROM goods_inventories), 1));
SELECT setval('goods_images_id_seq', GREATEST((SELECT MAX(id) FROM goods_images), 1));
