#!/bin/bash

set -e

BASE_URL="http://localhost:8081/papaya-payout-engine/v1"

echo "=== Papaya Payout Engine - Test Flow ==="
echo ""

echo "1. Health Check"
curl -s "$BASE_URL/../health-check" | jq .
echo ""

echo "2. Seed 150 test merchants"
SEED_RESULT=$(curl -s -X POST "$BASE_URL/merchants/seed" -H "Content-Type: application/json" -d '{"count": 150}')
echo "$SEED_RESULT" | jq .
echo ""

echo "3. List merchants (first 5)"
curl -s "$BASE_URL/merchants?limit=5" | jq '.merchants[] | {id, merchant_name, industry, chargeback_rate, account_age_days}'
echo ""

echo "4. Find and evaluate a high-risk merchant"
HIGH_RISK=$(curl -s "$BASE_URL/merchants?limit=150" | jq -r '.merchants[] | select((.chargeback_rate | tonumber) > 2) | .id' | head -1)
echo "High-risk merchant ID: $HIGH_RISK"
curl -s -X POST "$BASE_URL/risk/evaluate" -H "Content-Type: application/json" -d "{\"merchant_id\": \"$HIGH_RISK\", \"simulation\": false}" | jq '{risk_score, risk_level, payout_hold_period, rolling_reserve_percentage, reasoning: .reasoning.primary_factors}'
echo ""

echo "5. Find and evaluate a low-risk merchant"
LOW_RISK=$(curl -s "$BASE_URL/merchants?limit=150" | jq -r '.merchants[] | select((.chargeback_rate | tonumber) < 0.5 and .account_age_days > 365) | .id' | head -1)
echo "Low-risk merchant ID: $LOW_RISK"
curl -s -X POST "$BASE_URL/risk/evaluate" -H "Content-Type: application/json" -d "{\"merchant_id\": \"$LOW_RISK\", \"simulation\": false}" | jq '{risk_score, risk_level, payout_hold_period, rolling_reserve_percentage}'
echo ""

echo "6. Simulate risk improvement for high-risk merchant"
echo "Original risk score was HIGH/CRITICAL, now simulating with improved metrics..."
curl -s -X POST "$BASE_URL/risk/simulate" -H "Content-Type: application/json" -d "{\"merchant_id\": \"$HIGH_RISK\", \"overrides\": {\"chargeback_rate\": 0.4, \"account_age_days\": 800, \"kyc_verified\": true}}" | jq '{risk_score, risk_level, payout_hold_period, rolling_reserve_percentage, simulation}'
echo ""

echo "7. Batch evaluate 50 merchants"
MERCHANT_IDS=$(curl -s "$BASE_URL/merchants?limit=50" | jq -c '[.merchants[].id]')
curl -s -X POST "$BASE_URL/risk/batch-evaluate" -H "Content-Type: application/json" -d "{\"merchant_ids\": $MERCHANT_IDS, \"simulation\": false}" | jq '{total_merchants, summary, high_risk_count: (.high_risk_merchants | length), sample_high_risk: .high_risk_merchants[0]}'
echo ""

echo "8. Verify repeatability - evaluate same merchant twice"
echo "First evaluation:"
FIRST=$(curl -s -X POST "$BASE_URL/risk/evaluate" -H "Content-Type: application/json" -d "{\"merchant_id\": \"$LOW_RISK\", \"simulation\": true}" | jq '{risk_score, payout_hold_period, rolling_reserve_percentage}')
echo "$FIRST"

echo "Second evaluation:"
SECOND=$(curl -s -X POST "$BASE_URL/risk/evaluate" -H "Content-Type: application/json" -d "{\"merchant_id\": \"$LOW_RISK\", \"simulation\": true}" | jq '{risk_score, payout_hold_period, rolling_reserve_percentage}')
echo "$SECOND"

if [ "$FIRST" = "$SECOND" ]; then
    echo "✓ Repeatability verified: Same input produces same output"
else
    echo "✗ Repeatability failed: Different outputs for same input"
fi

echo ""
echo "=== Test Flow Complete ==="
