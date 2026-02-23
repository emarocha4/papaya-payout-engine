CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_name VARCHAR(255) NOT NULL,
    industry VARCHAR(50) NOT NULL,
    country VARCHAR(2) NOT NULL,

    transaction_volume_30d DECIMAL(15, 2) NOT NULL DEFAULT 0,
    transaction_count_30d INTEGER NOT NULL DEFAULT 0,
    avg_ticket_size DECIMAL(10, 2) NOT NULL DEFAULT 0,

    chargeback_count_30d INTEGER NOT NULL DEFAULT 0,
    chargeback_rate DECIMAL(5, 2) NOT NULL DEFAULT 0,
    refund_rate DECIMAL(5, 2) NOT NULL DEFAULT 0,
    velocity_multiplier DECIMAL(5, 2) NOT NULL DEFAULT 1.0,

    account_age_days INTEGER NOT NULL DEFAULT 0,
    account_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    kyc_verified BOOLEAN NOT NULL DEFAULT false,
    kyc_level VARCHAR(20) NOT NULL DEFAULT 'NONE',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chargeback_rate_valid CHECK (chargeback_rate >= 0 AND chargeback_rate <= 100)
);

CREATE INDEX idx_merchants_chargeback_rate ON merchants(chargeback_rate);
CREATE INDEX idx_merchants_account_age ON merchants(account_age_days);
CREATE INDEX idx_merchants_industry ON merchants(industry);
