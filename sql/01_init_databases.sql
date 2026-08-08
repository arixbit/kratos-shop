-- 初始化 kratos-shop 使用的数据库
-- 说明：PostgreSQL 的 CREATE DATABASE 不支持 IF NOT EXISTS，
-- 这里使用 psql 的 \gexec 实现幂等创建。

SELECT 'CREATE DATABASE shop_user'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop_user')\gexec

SELECT 'CREATE DATABASE shop_goods'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop_goods')\gexec

SELECT 'CREATE DATABASE shop_cart'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop_cart')\gexec

SELECT 'CREATE DATABASE shop_order'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop_order')\gexec

SELECT 'CREATE DATABASE shop_inventory'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop_inventory')\gexec

SELECT 'CREATE DATABASE shop_payment'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop_payment')\gexec
