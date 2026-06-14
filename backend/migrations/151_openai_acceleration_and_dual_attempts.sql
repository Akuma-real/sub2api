-- Per API-key OpenAI acceleration settings and dual-request attempt ledger.

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS acceleration_settings JSONB NOT NULL DEFAULT '{
        "fast_mode": "off",
        "dual_protection_enabled": false,
        "dual_first_response_timeout_ms": 8000
    }'::jsonb;

COMMENT ON COLUMN api_keys.acceleration_settings IS
    'Per-key OpenAI acceleration settings. Fast and dual protection are opt-in and default off.';

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS dual_protection_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS dual_attempt_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS dual_extra_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_breakdown JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN usage_logs.dual_protection_enabled IS 'Whether this request used API-key dual-request protection.';
COMMENT ON COLUMN usage_logs.dual_attempt_count IS 'Number of upstream attempts dispatched for dual-request protection.';
COMMENT ON COLUMN usage_logs.dual_extra_cost IS 'Additional billed cost from non-winning dual-request attempts.';
COMMENT ON COLUMN usage_logs.cost_breakdown IS 'Billing snapshot shown to users/admins, including Fast, dual protection, VIP, and final charge.';

CREATE TABLE IF NOT EXISTS openai_dual_attempts (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(128) NOT NULL,
    attempt_id VARCHAR(64) NOT NULL,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    endpoint VARCHAR(128) NOT NULL,
    method VARCHAR(16) NOT NULL DEFAULT 'POST',
    role VARCHAR(16) NOT NULL,
    outcome VARCHAR(16) NOT NULL DEFAULT 'pending',
    service_tier VARCHAR(32),
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    billing_basis VARCHAR(64),
    estimated_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    actual_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    billed_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    upstream_dispatched_at TIMESTAMPTZ,
    cancel_reason VARCHAR(128),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS openai_dual_attempts_request_api_key_attempt_idx
    ON openai_dual_attempts(request_id, api_key_id, attempt_id);

CREATE INDEX IF NOT EXISTS openai_dual_attempts_request_api_key_idx
    ON openai_dual_attempts(request_id, api_key_id);

CREATE INDEX IF NOT EXISTS openai_dual_attempts_api_key_created_idx
    ON openai_dual_attempts(api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS openai_dual_attempts_outcome_created_idx
    ON openai_dual_attempts(outcome, created_at DESC);

DROP INDEX IF EXISTS idx_usage_billing_dedup_request_api_key;

ALTER TABLE usage_billing_dedup
    ADD COLUMN IF NOT EXISTS attempt_id VARCHAR(64) NOT NULL DEFAULT 'primary';

ALTER TABLE usage_billing_dedup_archive
    ADD COLUMN IF NOT EXISTS attempt_id VARCHAR(64) NOT NULL DEFAULT 'primary';

CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_billing_dedup_request_api_key_attempt
    ON usage_billing_dedup (request_id, api_key_id, attempt_id);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'usage_billing_dedup_archive_pkey'
            AND conrelid = 'usage_billing_dedup_archive'::regclass
    ) THEN
        ALTER TABLE usage_billing_dedup_archive DROP CONSTRAINT usage_billing_dedup_archive_pkey;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS usage_billing_dedup_archive_request_api_key_attempt_idx
    ON usage_billing_dedup_archive (request_id, api_key_id, attempt_id);
