CREATE TABLE risk_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    batch_id UUID NULL,

    risk_score INTEGER NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    payout_hold_period VARCHAR(20) NOT NULL,
    rolling_reserve_percentage INTEGER NOT NULL,

    reasoning JSONB NOT NULL,

    simulation BOOLEAN NOT NULL DEFAULT false,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT risk_score_valid CHECK (risk_score >= 0 AND risk_score <= 100)
);

CREATE INDEX idx_decisions_merchant_id ON risk_decisions(merchant_id);
CREATE INDEX idx_decisions_evaluated_at ON risk_decisions(evaluated_at DESC);
CREATE INDEX idx_decisions_merchant_latest ON risk_decisions(merchant_id, evaluated_at DESC);
CREATE INDEX idx_decisions_batch_id ON risk_decisions(batch_id);
