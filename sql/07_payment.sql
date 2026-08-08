\connect shop_payment

CREATE TABLE IF NOT EXISTS payments (
    id          BIGSERIAL PRIMARY KEY,
    payment_no  VARCHAR(64)  NOT NULL UNIQUE,
    order_sn    VARCHAR(64)  NOT NULL,
    user_id     BIGINT       NOT NULL DEFAULT 0,
    amount      BIGINT       NOT NULL DEFAULT 0,
    channel     VARCHAR(32)  NOT NULL DEFAULT 'mock',
    status      INT          NOT NULL DEFAULT 1,
    trade_no    VARCHAR(100) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    paid_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payments_order ON payments (order_sn);

INSERT INTO payments (id, payment_no, order_sn, user_id, amount, channel, status)
VALUES (1, 'PAY-MOCK-00000001', '20260808013938373717', 1, 849900, 'mock', 2)
ON CONFLICT (id) DO NOTHING;

SELECT setval('payments_id_seq', GREATEST((SELECT MAX(id) FROM payments), 1));
