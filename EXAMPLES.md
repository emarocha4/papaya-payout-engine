# API Response Examples

## 1. Health Check

```bash
curl http://localhost:8081/health-check
```

```json
{
  "status": "OK",
  "database": "connected",
  "timestamp": "2026-02-23T11:00:00Z"
}
```

## 2. Seed Merchants

```bash
curl -X POST http://localhost:8081/papaya-payout-engine/v1/merchants/seed \
  -H "Content-Type: application/json" \
  -d '{"count": 150}'
```

```json
{
  "merchants_created": 150,
  "distribution": {
    "HIGH": 31,
    "LOW": 90,
    "MEDIUM": 29
  }
}
```

## 3. List Merchants

```bash
curl http://localhost:8081/papaya-payout-engine/v1/merchants?limit=3
```

```json
{
  "merchants": [
    {
      "id": "247efb61-8a85-43d5-8987-dd6118260c86",
      "merchant_name": "Pro Downloads",
      "industry": "DIGITAL_GOODS",
      "country": "CL",
      "transaction_volume_30d": "15320.45",
      "transaction_count_30d": 89,
      "avg_ticket_size": "172.14",
      "chargeback_count_30d": 4,
      "chargeback_rate": "4.49",
      "refund_rate": "8.23",
      "velocity_multiplier": "5.20",
      "account_age_days": 371,
      "account_created_at": "2025-02-17T00:00:00Z",
      "kyc_verified": false,
      "kyc_level": "NONE",
      "created_at": "2026-02-23T10:59:58Z",
      "updated_at": "2026-02-23T10:59:58Z"
    }
  ],
  "total": 150,
  "limit": 3,
  "offset": 0
}
```

## 4. Evaluate High-Risk Merchant

```bash
curl -X POST http://localhost:8081/papaya-payout-engine/v1/risk/evaluate \
  -H "Content-Type: application/json" \
  -d '{"merchant_id": "247efb61-8a85-43d5-8987-dd6118260c86", "simulation": false}'
```

```json
{
  "decision_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "merchant_id": "247efb61-8a85-43d5-8987-dd6118260c86",
  "batch_id": null,
  "risk_score": 75,
  "risk_level": "HIGH",
  "payout_hold_period": "45_DAYS",
  "rolling_reserve_percentage": 20,
  "reasoning": {
    "primary_factors": [
      {
        "factor": "Chargeback Rate",
        "score": 30,
        "contribution": "4.49% rate - Critical",
        "impact": "CRITICAL"
      },
      {
        "factor": "Account Age",
        "score": 5,
        "contribution": "Account 371 days old - Mature",
        "impact": "POSITIVE"
      },
      {
        "factor": "Transaction Velocity",
        "score": 15,
        "contribution": "5.2x velocity - High risk",
        "impact": "NEGATIVE"
      },
      {
        "factor": "Business Category",
        "score": 15,
        "contribution": "DIGITAL_GOODS - High risk category",
        "impact": "NEGATIVE"
      },
      {
        "factor": "KYC Verification",
        "score": 10,
        "contribution": "No KYC verification",
        "impact": "CRITICAL"
      }
    ],
    "policy_explanation": "Score of 75 places merchant in HIGH tier requiring 45_DAYS hold and 20% reserve"
  },
  "evaluated_at": "2026-02-23T11:00:38Z",
  "simulation": false
}
```

## 5. Evaluate Low-Risk Merchant

```bash
curl -X POST http://localhost:8081/papaya-payout-engine/v1/risk/evaluate \
  -H "Content-Type: application/json" \
  -d '{"merchant_id": "4edf3fa7-6ff5-4a3e-bd5a-bc651bbeba19", "simulation": false}'
```

```json
{
  "decision_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
  "merchant_id": "4edf3fa7-6ff5-4a3e-bd5a-bc651bbeba19",
  "risk_score": 33,
  "risk_level": "MEDIUM_LOW",
  "payout_hold_period": "7_DAYS",
  "rolling_reserve_percentage": 0,
  "reasoning": {
    "primary_factors": [
      {
        "factor": "Chargeback Rate",
        "score": 0,
        "contribution": "0.22% rate - Excellent",
        "impact": "POSITIVE"
      },
      {
        "factor": "Account Age",
        "score": 10,
        "contribution": "Account 350 days old - Established",
        "impact": "NEUTRAL"
      },
      {
        "factor": "Transaction Velocity",
        "score": 5,
        "contribution": "1.8x velocity - Elevated",
        "impact": "NEUTRAL"
      },
      {
        "factor": "Business Category",
        "score": 15,
        "contribution": "ELECTRONICS - High risk category",
        "impact": "NEGATIVE"
      },
      {
        "factor": "KYC Verification",
        "score": 3,
        "contribution": "Full KYC - ID and address verified",
        "impact": "NEUTRAL"
      }
    ],
    "policy_explanation": "Score of 33 places merchant in MEDIUM_LOW tier requiring 7_DAYS hold and 0% reserve"
  },
  "evaluated_at": "2026-02-23T11:00:40Z",
  "simulation": false
}
```

## 6. Get Merchant Profile

```bash
curl http://localhost:8081/papaya-payout-engine/v1/risk/merchants/4edf3fa7-6ff5-4a3e-bd5a-bc651bbeba19/profile
```

```json
{
  "merchant_id": "4edf3fa7-6ff5-4a3e-bd5a-bc651bbeba19",
  "merchant_name": "Smart Digital",
  "industry": "ELECTRONICS",
  "country": "CO",
  "account_created_at": "2025-03-10T10:59:58Z",
  "account_age_days": 350,
  "risk_metrics": {
    "transaction_volume_30d": "53325.17",
    "transaction_count_30d": 458,
    "avg_ticket_size": "116.43",
    "chargeback_count_30d": 1,
    "chargeback_rate": "0.22",
    "refund_rate": "5.87",
    "velocity_multiplier": "1.84",
    "kyc_verified": true,
    "kyc_level": "FULL"
  },
  "current_policy": {
    "risk_score": 33,
    "payout_hold_period": "7_DAYS",
    "rolling_reserve_percentage": 0,
    "last_evaluated_at": "2026-02-23T11:00:40Z"
  }
}
```

## 7. Simulate Risk Improvement

```bash
curl -X POST http://localhost:8081/papaya-payout-engine/v1/risk/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "247efb61-8a85-43d5-8987-dd6118260c86",
    "overrides": {
      "chargeback_rate": 0.3,
      "account_age_days": 800,
      "kyc_verified": true
    }
  }'
```

```json
{
  "decision_id": "00000000-0000-0000-0000-000000000000",
  "merchant_id": "247efb61-8a85-43d5-8987-dd6118260c86",
  "risk_score": 33,
  "risk_level": "MEDIUM_LOW",
  "payout_hold_period": "7_DAYS",
  "rolling_reserve_percentage": 0,
  "reasoning": {
    "primary_factors": [
      {
        "factor": "Chargeback Rate",
        "score": 0,
        "contribution": "0.30% rate - Excellent",
        "impact": "POSITIVE"
      },
      {
        "factor": "Account Age",
        "score": 0,
        "contribution": "Account 800 days old - Veteran",
        "impact": "POSITIVE"
      },
      {
        "factor": "Transaction Velocity",
        "score": 15,
        "contribution": "5.2x velocity - High risk",
        "impact": "NEGATIVE"
      },
      {
        "factor": "Business Category",
        "score": 15,
        "contribution": "DIGITAL_GOODS - High risk category",
        "impact": "NEGATIVE"
      },
      {
        "factor": "KYC Verification",
        "score": 3,
        "contribution": "Full KYC - ID and address verified",
        "impact": "NEUTRAL"
      }
    ],
    "policy_explanation": "Score of 33 places merchant in MEDIUM_LOW tier requiring 7_DAYS hold and 0% reserve"
  },
  "evaluated_at": "2026-02-23T11:01:00Z",
  "simulation": true
}
```

**Note**: With improved chargeback rate (0.3%) and account age (800 days), the risk score dropped from 75 (HIGH) to 33 (MEDIUM_LOW), improving payout terms from 45 days to 7 days.

## 8. Batch Evaluate

```bash
curl -X POST http://localhost:8081/papaya-payout-engine/v1/risk/batch-evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_ids": ["247efb61-8a85-43d5-8987-dd6118260c86", "4edf3fa7-6ff5-4a3e-bd5a-bc651bbeba19", ...],
    "simulation": false
  }'
```

```json
{
  "batch_id": "e1f2a3b4-c5d6-7890-abcd-ef1234567890",
  "total_merchants": 30,
  "evaluated_at": "2026-02-23T11:02:00Z",
  "summary": {
    "by_hold_period": {
      "IMMEDIATE": 2,
      "7_DAYS": 4,
      "14_DAYS": 2,
      "45_DAYS": 22
    },
    "by_reserve": {
      "0_PERCENT": 6,
      "10_PERCENT": 2,
      "20_PERCENT": 22
    },
    "by_risk_level": {
      "LOW": 2,
      "MEDIUM_LOW": 4,
      "MEDIUM": 2,
      "HIGH": 11,
      "CRITICAL": 11
    }
  },
  "high_risk_merchants": [
    {
      "merchant_id": "247efb61-8a85-43d5-8987-dd6118260c86",
      "risk_score": 75,
      "risk_level": "HIGH",
      "primary_concerns": [
        "4.49% rate - Critical",
        "5.2x velocity - High risk",
        "DIGITAL_GOODS - High risk category",
        "No KYC verification"
      ],
      "recommended_action": "Manual review required"
    },
    {
      "merchant_id": "abc12345-6789-0abc-def1-234567890abc",
      "risk_score": 85,
      "risk_level": "CRITICAL",
      "primary_concerns": [
        "6.20% rate - Critical",
        "Account 25 days old - Very new",
        "8.5x velocity - Critical",
        "TRAVEL - High risk category"
      ],
      "recommended_action": "Immediate manual approval required"
    }
  ],
  "simulation": false
}
```

## Key Observations

### Risk Score Distribution
- **LOW (0-20)**: Established merchants with < 0.5% chargeback rate
- **MEDIUM-LOW (21-40)**: Growing businesses with acceptable metrics
- **MEDIUM (41-60)**: Merchants requiring enhanced monitoring
- **HIGH (61-80)**: New or problematic merchants requiring review
- **CRITICAL (81-100)**: High-risk merchants requiring immediate attention

### Business Impact
- **Immediate payout**: Low-risk, trusted merchants (best terms)
- **7-day hold**: Standard processing for stable merchants
- **14-day hold + 10% reserve**: Enhanced monitoring for medium-risk
- **45-day hold + 20% reserve**: Maximum protection for high-risk

### Repeatability
Running the same evaluation twice on the same merchant (with simulation=true) produces identical risk scores, demonstrating the deterministic nature of the scoring model.
