-- Add VIP redemption target fields to redeem codes.
ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS vip_level_id BIGINT;
ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS vip_days INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_redeem_codes_vip_level_id ON redeem_codes(vip_level_id);
