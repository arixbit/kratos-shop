
-- =====================================================================
-- 用户表
-- =====================================================================
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    mobile        VARCHAR(11)  NOT NULL UNIQUE,
    password      VARCHAR(200) NOT NULL,
    nick_name     VARCHAR(25)  NOT NULL DEFAULT '',
    birthday      TIMESTAMPTZ,
    gender        VARCHAR(16)  NOT NULL DEFAULT 'male',
    role          INT          NOT NULL DEFAULT 1,
    add_time      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    is_deleted_at BOOLEAN      NOT NULL DEFAULT FALSE
);

COMMENT ON COLUMN users.mobile IS '手机号码，用户唯一标识';
COMMENT ON COLUMN users.gender IS 'female: 女, male: 男';
COMMENT ON COLUMN users.role IS '1: 普通用户, 2: 管理员';

-- =====================================================================
-- 用户收货地址表
-- =====================================================================
CREATE TABLE IF NOT EXISTS user_addresses (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT       NOT NULL DEFAULT 1,
    is_default SMALLINT     NOT NULL DEFAULT 0,
    mobile     VARCHAR(11)  NOT NULL,
    name       VARCHAR(25)  NOT NULL DEFAULT '',
    province   VARCHAR(25)  NOT NULL DEFAULT '',
    city       VARCHAR(25)  NOT NULL DEFAULT '',
    districts  VARCHAR(25)  NOT NULL DEFAULT '',
    address    VARCHAR(255) NOT NULL DEFAULT '',
    post_code  VARCHAR(25)  NOT NULL DEFAULT '',
    add_time   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON user_addresses (user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_mobile ON user_addresses (mobile);

-- =====================================================================
-- Mock 基础数据
-- =====================================================================
-- 密码均为 12345678
INSERT INTO users (id, mobile, password, nick_name, birthday, gender, role, add_time, update_time)
VALUES
    (1, '13800138000', '$pbkdf2-sha512$781f496fd9233c152837110ac2ce8a79$0408f25f5393afa7259e9fad0fceb1a2f3e406641d221b0b5562fb9ad46f8406',
     '商城用户', '1995-06-15 00:00:00+08', 'male', 1, NOW(), NOW()),
    (2, '13501167215', '$pbkdf2-sha512$781f496fd9233c152837110ac2ce8a79$0408f25f5393afa7259e9fad0fceb1a2f3e406641d221b0b5562fb9ad46f8406',
     '管理员', '1990-01-01 00:00:00+08', 'male', 2, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_addresses (id, user_id, is_default, mobile, name, province, city, districts, address, post_code, add_time, update_time)
VALUES
    (1, 1, 1, '13800138000', '张三', '广东省', '深圳市', '南山区', '科技园路 1 号', '518000', NOW(), NOW()),
    (2, 1, 0, '13800138000', '张三', '广东省', '广州市', '天河区', '体育西路 100 号', '510000', NOW(), NOW()),
    (3, 2, 1, '13501167215', '李四', '上海市', '上海市', '浦东新区', '世纪大道 2000 号', '200000', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

SELECT setval('users_id_seq', GREATEST((SELECT MAX(id) FROM users), 1));
SELECT setval('user_addresses_id_seq', GREATEST((SELECT MAX(id) FROM user_addresses), 1));
