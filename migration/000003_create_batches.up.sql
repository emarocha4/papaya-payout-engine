CREATE TABLE batch_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    total_merchants INTEGER NOT NULL,
    summary JSONB NOT NULL,
    high_risk_merchants JSONB NOT NULL,
    simulation BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batch_reports_created_at ON batch_reports(created_at DESC);
