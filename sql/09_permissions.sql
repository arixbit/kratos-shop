\connect shop_user

-- =====================================================================
-- 权限点表
-- =====================================================================
CREATE TABLE IF NOT EXISTS permissions (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(100) NOT NULL UNIQUE,
    name       VARCHAR(100) NOT NULL DEFAULT '',
    group_name VARCHAR(100) NOT NULL DEFAULT '',
    sort       INT          NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- 角色权限关联表（role 沿用 users.role：1 普通用户 / 2 管理员）
-- =====================================================================
CREATE TABLE IF NOT EXISTS role_permissions (
    id            BIGSERIAL PRIMARY KEY,
    role_id       INT  NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_role_permission UNIQUE (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions (role_id);

-- =====================================================================
-- 默认权限点
-- =====================================================================
INSERT INTO permissions (code, name, group_name, sort)
VALUES
    ('goods:create', '新增商品', '商品管理', 10),
    ('goods:update', '编辑商品', '商品管理', 20),
    ('goods:delete', '删除商品', '商品管理', 30),
    ('goods:status', '上下架商品', '商品管理', 40),
    ('goods:category:manage', '商品分类管理', '商品管理', 50),
    ('order:ship', '订单发货', '订单管理', 10),
    ('order:refund', '订单退款', '订单管理', 20),
    ('user:address:delete', '删除用户地址', '用户管理', 10),
    ('system:permission:manage', '权限管理', '系统管理', 10)
ON CONFLICT (code) DO NOTHING;

-- 管理员默认拥有全部权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;
