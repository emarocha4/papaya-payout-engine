# Papaya Commerce Risk-Based Payout Engine

Backend service in Go that evaluates merchant risk profiles and determines intelligent payout policies.

## Quick Start

**Easiest way (one command):**
```bash
./start.sh
```

**Manual setup:**
```bash
# 1. Ensure PostgreSQL is running
docker ps | grep postgres

# 2. Build and start
go mod download
go build -o papaya-payout-engine main.go
DB_USER=postgres DB_PASSWORD=postgres PORT=8081 ./papaya-payout-engine
```

Server runs on `http://localhost:8081`

## Automated Test Flow

```bash
# Run comprehensive test flow
./test_flow.sh
```

This script will:
1. Check health
2. Seed 150 test merchants
3. Evaluate high-risk and low-risk merchants
4. Simulate risk improvements
5. Batch evaluate multiple merchants
6. Verify repeatability

## API Endpoints

### 1. Seed Test Data
```bash
curl -X POST http://localhost:8080/papaya-payout-engine/v1/merchants/seed \
  -H "Content-Type: application/json" \
  -d '{"count": 150}'
```

### 2. List Merchants
```bash
curl http://localhost:8080/papaya-payout-engine/v1/merchants?limit=10
```

### 3. Evaluate Risk
```bash
curl -X POST http://localhost:8080/papaya-payout-engine/v1/risk/evaluate \
  -H "Content-Type: application/json" \
  -d '{"merchant_id": "YOUR_MERCHANT_ID", "simulation": false}'
```

### 4. Get Merchant Profile
```bash
curl http://localhost:8080/papaya-payout-engine/v1/risk/merchants/YOUR_MERCHANT_ID/profile
```

### 5. Simulate Risk Changes

Simulate with merchant data overrides:
```bash
curl -X POST http://localhost:8080/papaya-payout-engine/v1/risk/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "YOUR_MERCHANT_ID",
    "overrides": {
      "chargeback_rate": 0.5,
      "account_age_days": 400,
      "kyc_verified": true
    }
  }'
```

Simulate with custom scoring thresholds:
```bash
curl -X POST http://localhost:8080/papaya-payout-engine/v1/risk/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "YOUR_MERCHANT_ID",
    "overrides": {
      "scoring_thresholds": {
        "chargeback_excellent": 0.3,
        "chargeback_critical": 1.0,
        "velocity_normal": 1.2,
        "refund_normal": 2.5
      }
    }
  }'
```

### 6. Batch Evaluate
```bash
curl -X POST http://localhost:8080/papaya-payout-engine/v1/risk/batch-evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_ids": ["id1", "id2", "id3"],
    "simulation": false
  }'
```

### 7. Health Check
```bash
curl http://localhost:8080/health-check
```

## Risk Scoring Model

### Factors (100 points total)
- **Chargeback Rate** (30 points): < 0.5% = 0pts, 0.5-1% = 10pts, 1-1.5% = 20pts, > 1.5% = 30pts
- **Account Age** (25 points): < 30 days = 25pts, decreasing to 0pts for > 730 days
- **Transaction Velocity** (20 points): < 1.5x = 0pts, increasing to 20pts for > 6x
- **Business Category** (15 points): DIGITAL_GOODS/TRAVEL/ELECTRONICS = 15pts, UTILITIES/HEALTHCARE = 0pts
- **KYC Verification** (10 points): No KYC = 10pts, ENHANCED = 0pts
- **Refund Rate** (5 points): < 3% = 0pts, 3-6% = 3pts, > 6% = 5pts (fraud signal)

### Policy Tiers
- **0-20 (LOW)**: IMMEDIATE payout, 0% reserve
- **21-40 (MEDIUM-LOW)**: 7_DAYS hold, 0% reserve
- **41-60 (MEDIUM)**: 14_DAYS hold, 10% reserve
- **61-80 (HIGH)**: 45_DAYS hold, 20% reserve
- **81-100 (CRITICAL)**: 45_DAYS hold, 20% reserve

## Features

### Core Capabilities
- ✅ **Real-time Risk Assessment**: Evaluate merchants in <100ms
- ✅ **Batch Processing**: Process up to 100 merchants concurrently
- ✅ **Simulation Mode**: Test "what-if" scenarios without persisting data
- ✅ **Configurable Thresholds**: Override scoring rules for custom risk policies
- ✅ **Comprehensive Logging**: Structured logs for debugging and audit trails
- ✅ **Error Tracking**: Detailed error messages with context for troubleshooting

### Testing & Quality
- ✅ **69 Unit Tests**: Comprehensive test coverage for core logic
- ✅ **Table-Driven Tests**: Edge cases and boundary validation
- ✅ **Integration Tests**: HTTP layer testing with mocks
- ✅ **Race Condition Testing**: Concurrent operation safety

### Performance & Reliability
- ✅ **Context Timeouts**: 30s batch timeout, 10s service timeout
- ✅ **Input Validation**: Max 100 merchants per batch, UUID validation
- ✅ **Concurrent Processing**: Goroutine pools with semaphores
- ✅ **Graceful Error Handling**: Partial success reporting in batch operations

## Project Structure

```
papaya-payout-engine/
├── cmd/server/           # HTTP server and handlers
├── internal/
│   ├── risk/            # Risk evaluation engine
│   ├── merchant/        # Merchant domain
│   ├── store/           # Data persistence
│   ├── platform/        # Infrastructure
│   └── health/          # Health checks
├── migration/           # SQL migrations
└── main.go             # Entry point
```

## Tech Stack

- Go 1.24+
- PostgreSQL 14
- Echo v4 (HTTP framework)
- GORM (ORM)
- shopspring/decimal (financial precision)

## Environment Variables

```
ENVIRONMENT=local
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=papaya_user
DB_PASSWORD=papaya_pass
DB_NAME=papaya_payout_engine
DB_SSLMODE=disable
```

## Testing Flow

1. Seed 150 test merchants (60% LOW, 25% MEDIUM, 15% HIGH risk)
2. Evaluate individual merchants and verify risk scores
3. Test simulation mode with different parameters
4. Run batch evaluation on all merchants
5. Verify repeatability (same input = same output)
