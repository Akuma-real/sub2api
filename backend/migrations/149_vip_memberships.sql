-- VIP levels and memberships.
CREATE TABLE IF NOT EXISTS vip_levels (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price DECIMAL(20,2) NOT NULL,
    original_price DECIMAL(20,2) NULL,
    validity_days INTEGER NOT NULL DEFAULT 30,
    discount_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1,
    features TEXT NOT NULL DEFAULT '',
    benefits JSONB NOT NULL DEFAULT '{}'::jsonb,
    for_sale BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vip_levels_price_positive CHECK (price > 0),
    CONSTRAINT vip_levels_original_price_nonnegative CHECK (original_price IS NULL OR original_price >= 0),
    CONSTRAINT vip_levels_validity_days_positive CHECK (validity_days > 0),
    CONSTRAINT vip_levels_discount_multiplier_range CHECK (discount_multiplier > 0 AND discount_multiplier <= 1)
);

CREATE INDEX IF NOT EXISTS vip_levels_for_sale_idx ON vip_levels (for_sale);
CREATE INDEX IF NOT EXISTS vip_levels_sort_order_idx ON vip_levels (sort_order);

CREATE TABLE IF NOT EXISTS user_vip_memberships (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vip_level_id BIGINT NOT NULL REFERENCES vip_levels(id) ON DELETE RESTRICT,
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    source_order_id BIGINT NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_vip_memberships_period_valid CHECK (expires_at > starts_at)
);

CREATE INDEX IF NOT EXISTS user_vip_memberships_user_id_idx ON user_vip_memberships (user_id);
CREATE INDEX IF NOT EXISTS user_vip_memberships_vip_level_id_idx ON user_vip_memberships (vip_level_id);
CREATE INDEX IF NOT EXISTS user_vip_memberships_status_idx ON user_vip_memberships (status);
CREATE INDEX IF NOT EXISTS user_vip_memberships_expires_at_idx ON user_vip_memberships (expires_at);
CREATE INDEX IF NOT EXISTS user_vip_memberships_source_order_id_idx ON user_vip_memberships (source_order_id);
CREATE INDEX IF NOT EXISTS user_vip_memberships_user_status_expires_idx ON user_vip_memberships (user_id, status, expires_at);

ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS vip_level_id BIGINT NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS vip_days INTEGER NULL;
CREATE INDEX IF NOT EXISTS payment_orders_vip_level_id_idx ON payment_orders (vip_level_id);

ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS vip_level_id BIGINT NULL;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS vip_discount_multiplier DECIMAL(10,4) NULL;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS vip_pre_discount_cost DECIMAL(20,10) NULL;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS vip_savings_usd DECIMAL(20,10) NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS usage_logs_vip_level_id_idx ON usage_logs (vip_level_id);
