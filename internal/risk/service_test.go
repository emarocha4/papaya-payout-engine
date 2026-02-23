package risk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
)

type mockMerchantRepository struct {
	getMerchant func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error)
}

func (m *mockMerchantRepository) Get(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
	if m.getMerchant != nil {
		return m.getMerchant(ctx, id)
	}
	return nil, errors.New("not implemented")
}

type mockDecisionRepository struct {
	createDecision          func(ctx context.Context, decision *RiskDecision) error
	getLatestByMerchant     func(ctx context.Context, merchantID uuid.UUID) (*RiskDecision, error)
	bulkCreate              func(ctx context.Context, decisions []RiskDecision) error
}

func (m *mockDecisionRepository) Create(ctx context.Context, decision *RiskDecision) error {
	if m.createDecision != nil {
		return m.createDecision(ctx, decision)
	}
	return nil
}

func (m *mockDecisionRepository) GetLatestByMerchant(ctx context.Context, merchantID uuid.UUID) (*RiskDecision, error) {
	if m.getLatestByMerchant != nil {
		return m.getLatestByMerchant(ctx, merchantID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDecisionRepository) BulkCreate(ctx context.Context, decisions []RiskDecision) error {
	if m.bulkCreate != nil {
		return m.bulkCreate(ctx, decisions)
	}
	return nil
}

func TestEvaluateMerchant(t *testing.T) {
	merchantID := uuid.New()
	testMerchant := &merchant.Merchant{
		ID:                 merchantID,
		MerchantName:       "Test Merchant",
		Industry:           "RETAIL",
		Country:            "US",
		AccountAgeDays:     800,
		ChargebackRate:     decimal.NewFromFloat(0.3),
		RefundRate:         decimal.NewFromFloat(2.0),
		VelocityMultiplier: decimal.NewFromFloat(1.2),
		KYCVerified:        true,
		KYCLevel:           "ENHANCED",
	}

	t.Run("successful evaluation - no simulation", func(t *testing.T) {
		var savedDecision *RiskDecision
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				if id != merchantID {
					t.Errorf("expected merchant ID %v, got %v", merchantID, id)
				}
				return testMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			createDecision: func(ctx context.Context, decision *RiskDecision) error {
				savedDecision = decision
				return nil
			},
		}

		service := NewService(merchantStore, decisionStore)
		decision, err := service.EvaluateMerchant(context.Background(), merchantID, false)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decision == nil {
			t.Fatal("expected decision, got nil")
		}
		if decision.MerchantID != merchantID {
			t.Errorf("expected merchant ID %v, got %v", merchantID, decision.MerchantID)
		}
		if decision.Simulation {
			t.Error("expected simulation=false")
		}
		if savedDecision == nil {
			t.Error("expected decision to be saved, but it wasn't")
		}
		if decision.RiskScore != 5 {
			t.Errorf("expected risk score 5 (only category risk), got %d", decision.RiskScore)
		}
	})

	t.Run("successful evaluation - with simulation", func(t *testing.T) {
		var savedDecision *RiskDecision
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return testMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			createDecision: func(ctx context.Context, decision *RiskDecision) error {
				savedDecision = decision
				return nil
			},
		}

		service := NewService(merchantStore, decisionStore)
		decision, err := service.EvaluateMerchant(context.Background(), merchantID, true)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !decision.Simulation {
			t.Error("expected simulation=true")
		}
		if savedDecision != nil {
			t.Error("expected decision NOT to be saved in simulation mode")
		}
	})

	t.Run("merchant not found", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return nil, errors.New("merchant not found")
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		decision, err := service.EvaluateMerchant(context.Background(), merchantID, false)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if decision != nil {
			t.Error("expected nil decision when merchant not found")
		}
	})

	t.Run("decision save failure", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return testMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			createDecision: func(ctx context.Context, decision *RiskDecision) error {
				return errors.New("database error")
			},
		}

		service := NewService(merchantStore, decisionStore)
		decision, err := service.EvaluateMerchant(context.Background(), merchantID, false)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if decision != nil {
			t.Error("expected nil decision when save fails")
		}
	})

	t.Run("high risk merchant evaluation", func(t *testing.T) {
		highRiskMerchant := &merchant.Merchant{
			ID:                 merchantID,
			MerchantName:       "High Risk Merchant",
			Industry:           "DIGITAL_GOODS",
			AccountAgeDays:     15,
			ChargebackRate:     decimal.NewFromFloat(4.5),
			RefundRate:         decimal.NewFromFloat(8.5),
			VelocityMultiplier: decimal.NewFromFloat(8.0),
			KYCVerified:        false,
			KYCLevel:           "NONE",
		}

		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return highRiskMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			createDecision: func(ctx context.Context, decision *RiskDecision) error {
				return nil
			},
		}

		service := NewService(merchantStore, decisionStore)
		decision, err := service.EvaluateMerchant(context.Background(), merchantID, false)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decision.RiskScore != 100 {
			t.Errorf("expected max risk score 100, got %d", decision.RiskScore)
		}
		if decision.RiskLevel != RiskLevelCritical {
			t.Errorf("expected CRITICAL risk level, got %v", decision.RiskLevel)
		}
	})
}

func TestSimulateMerchant(t *testing.T) {
	merchantID := uuid.New()
	baseMerchant := &merchant.Merchant{
		ID:                 merchantID,
		MerchantName:       "Test Merchant",
		Industry:           "RETAIL",
		AccountAgeDays:     800,
		ChargebackRate:     decimal.NewFromFloat(0.3),
		RefundRate:         decimal.NewFromFloat(2.0),
		VelocityMultiplier: decimal.NewFromFloat(1.2),
		KYCVerified:        true,
		KYCLevel:           "ENHANCED",
	}

	t.Run("simulate with chargeback rate override", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return baseMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		overrides := map[string]interface{}{
			"chargeback_rate": 4.5,
		}

		decision, err := service.SimulateMerchant(context.Background(), merchantID, overrides)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !decision.Simulation {
			t.Error("expected simulation=true")
		}
		if decision.RiskScore <= 5 {
			t.Errorf("expected higher risk score with 4.5%% chargeback, got %d", decision.RiskScore)
		}
	})

	t.Run("simulate with account age override", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return baseMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		overrides := map[string]interface{}{
			"account_age_days": 15.0,
		}

		decision, err := service.SimulateMerchant(context.Background(), merchantID, overrides)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decision.RiskScore <= 20 {
			t.Errorf("expected score > 20 with 15 days age, got %d", decision.RiskScore)
		}
	})

	t.Run("simulate with velocity override", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return baseMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		overrides := map[string]interface{}{
			"velocity_multiplier": 8.0,
		}

		decision, err := service.SimulateMerchant(context.Background(), merchantID, overrides)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decision.RiskScore <= 20 {
			t.Errorf("expected score > 20 with 8x velocity, got %d", decision.RiskScore)
		}
	})

	t.Run("simulate with kyc override", func(t *testing.T) {
		unverifiedMerchant := &merchant.Merchant{
			ID:                 merchantID,
			MerchantName:       "Unverified Merchant",
			Industry:           "RETAIL",
			AccountAgeDays:     800,
			ChargebackRate:     decimal.NewFromFloat(0.3),
			RefundRate:         decimal.NewFromFloat(2.0),
			VelocityMultiplier: decimal.NewFromFloat(1.2),
			KYCVerified:        false,
			KYCLevel:           "NONE",
		}

		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return unverifiedMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)

		baseDecision, _ := service.SimulateMerchant(context.Background(), merchantID, map[string]interface{}{})
		baseScore := baseDecision.RiskScore

		overrides := map[string]interface{}{
			"kyc_verified": true,
		}
		decision, err := service.SimulateMerchant(context.Background(), merchantID, overrides)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decision.RiskScore >= baseScore {
			t.Errorf("expected lower score with KYC verified, base=%d, new=%d", baseScore, decision.RiskScore)
		}
	})

	t.Run("simulate with multiple overrides", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return baseMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		overrides := map[string]interface{}{
			"chargeback_rate":     4.5,
			"account_age_days":    15.0,
			"velocity_multiplier": 8.0,
		}

		decision, err := service.SimulateMerchant(context.Background(), merchantID, overrides)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decision.RiskScore < 50 {
			t.Errorf("expected high risk score with multiple negative overrides, got %d", decision.RiskScore)
		}
	})

	t.Run("merchant not found", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return nil, errors.New("merchant not found")
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		decision, err := service.SimulateMerchant(context.Background(), merchantID, map[string]interface{}{})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if decision != nil {
			t.Error("expected nil decision when merchant not found")
		}
	})
}

func TestGetMerchantProfile(t *testing.T) {
	merchantID := uuid.New()
	testMerchant := &merchant.Merchant{
		ID:                   merchantID,
		MerchantName:         "Test Merchant",
		Industry:             "RETAIL",
		Country:              "US",
		AccountCreatedAt:     time.Now().AddDate(0, 0, -800),
		AccountAgeDays:       800,
		TransactionVolume30d: decimal.NewFromFloat(100000),
		TransactionCount30d:  500,
		AvgTicketSize:        decimal.NewFromFloat(200),
		ChargebackCount30d:   15,
		ChargebackRate:       decimal.NewFromFloat(0.3),
		RefundRate:           decimal.NewFromFloat(2.0),
		VelocityMultiplier:   decimal.NewFromFloat(1.2),
		KYCVerified:          true,
		KYCLevel:             "ENHANCED",
	}

	testDecision := &RiskDecision{
		ID:                       uuid.New(),
		MerchantID:               merchantID,
		RiskScore:                15,
		RiskLevel:                RiskLevelLow,
		PayoutHoldPeriod:         HoldPeriodImmediate,
		RollingReservePercentage: 0,
		EvaluatedAt:              time.Now(),
	}

	t.Run("successful profile retrieval", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return testMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			getLatestByMerchant: func(ctx context.Context, merchantID uuid.UUID) (*RiskDecision, error) {
				return testDecision, nil
			},
		}

		service := NewService(merchantStore, decisionStore)
		profile, err := service.GetMerchantProfile(context.Background(), merchantID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile == nil {
			t.Fatal("expected profile, got nil")
		}
		if profile.MerchantID != merchantID {
			t.Errorf("expected merchant ID %v, got %v", merchantID, profile.MerchantID)
		}
		if profile.MerchantName != testMerchant.MerchantName {
			t.Errorf("expected merchant name %s, got %s", testMerchant.MerchantName, profile.MerchantName)
		}
		if profile.CurrentPolicy == nil {
			t.Fatal("expected current policy, got nil")
		}
		if profile.CurrentPolicy.RiskScore != testDecision.RiskScore {
			t.Errorf("expected risk score %d, got %d", testDecision.RiskScore, profile.CurrentPolicy.RiskScore)
		}
	})

	t.Run("profile without decision history", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return testMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			getLatestByMerchant: func(ctx context.Context, merchantID uuid.UUID) (*RiskDecision, error) {
				return nil, nil
			},
		}

		service := NewService(merchantStore, decisionStore)
		profile, err := service.GetMerchantProfile(context.Background(), merchantID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.CurrentPolicy != nil {
			t.Error("expected nil current policy for merchant without decision history")
		}
	})

	t.Run("merchant not found", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return nil, errors.New("merchant not found")
			},
		}
		decisionStore := &mockDecisionRepository{}

		service := NewService(merchantStore, decisionStore)
		profile, err := service.GetMerchantProfile(context.Background(), merchantID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if profile != nil {
			t.Error("expected nil profile when merchant not found")
		}
	})

	t.Run("decision fetch failure", func(t *testing.T) {
		merchantStore := &mockMerchantRepository{
			getMerchant: func(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
				return testMerchant, nil
			},
		}
		decisionStore := &mockDecisionRepository{
			getLatestByMerchant: func(ctx context.Context, merchantID uuid.UUID) (*RiskDecision, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(merchantStore, decisionStore)
		profile, err := service.GetMerchantProfile(context.Background(), merchantID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if profile != nil {
			t.Error("expected nil profile when decision fetch fails")
		}
	})
}

func TestApplyOverrides(t *testing.T) {
	service := NewService(nil, nil)

	t.Run("apply all overrides", func(t *testing.T) {
		m := &merchant.Merchant{
			ChargebackRate:     decimal.NewFromFloat(0.5),
			AccountAgeDays:     100,
			KYCVerified:        false,
			KYCLevel:           "NONE",
			VelocityMultiplier: decimal.NewFromFloat(1.0),
		}

		overrides := map[string]interface{}{
			"chargeback_rate":     2.5,
			"account_age_days":    500.0,
			"kyc_verified":        true,
			"velocity_multiplier": 3.0,
		}

		service.applyOverrides(m, overrides)

		if m.ChargebackRate.InexactFloat64() != 2.5 {
			t.Errorf("expected chargeback rate 2.5, got %v", m.ChargebackRate.InexactFloat64())
		}
		if m.AccountAgeDays != 500 {
			t.Errorf("expected account age 500, got %d", m.AccountAgeDays)
		}
		if !m.KYCVerified {
			t.Error("expected KYC verified to be true")
		}
		if m.KYCLevel != "FULL" {
			t.Errorf("expected KYC level FULL when verified, got %s", m.KYCLevel)
		}
		if m.VelocityMultiplier.InexactFloat64() != 3.0 {
			t.Errorf("expected velocity multiplier 3.0, got %v", m.VelocityMultiplier.InexactFloat64())
		}
	})

	t.Run("kyc_verified updates level", func(t *testing.T) {
		m := &merchant.Merchant{
			KYCVerified: false,
			KYCLevel:    "NONE",
		}

		overrides := map[string]interface{}{
			"kyc_verified": true,
		}

		service.applyOverrides(m, overrides)

		if !m.KYCVerified {
			t.Error("expected KYC verified to be true")
		}
		if m.KYCLevel != "FULL" {
			t.Errorf("expected KYC level to be updated to FULL, got %s", m.KYCLevel)
		}
	})

	t.Run("partial overrides", func(t *testing.T) {
		original := &merchant.Merchant{
			ChargebackRate:     decimal.NewFromFloat(0.5),
			AccountAgeDays:     100,
			VelocityMultiplier: decimal.NewFromFloat(1.0),
		}

		overrides := map[string]interface{}{
			"chargeback_rate": 1.5,
		}

		service.applyOverrides(original, overrides)

		if original.ChargebackRate.InexactFloat64() != 1.5 {
			t.Error("chargeback rate should be overridden")
		}
		if original.AccountAgeDays != 100 {
			t.Error("account age should not change")
		}
		if original.VelocityMultiplier.InexactFloat64() != 1.0 {
			t.Error("velocity multiplier should not change")
		}
	})

	t.Run("empty overrides", func(t *testing.T) {
		original := &merchant.Merchant{
			ChargebackRate:     decimal.NewFromFloat(0.5),
			AccountAgeDays:     100,
			VelocityMultiplier: decimal.NewFromFloat(1.0),
		}

		service.applyOverrides(original, map[string]interface{}{})

		if original.ChargebackRate.InexactFloat64() != 0.5 {
			t.Error("chargeback rate should not change")
		}
		if original.AccountAgeDays != 100 {
			t.Error("account age should not change")
		}
	})

	t.Run("invalid type overrides ignored", func(t *testing.T) {
		original := &merchant.Merchant{
			ChargebackRate: decimal.NewFromFloat(0.5),
			AccountAgeDays: 100,
		}

		overrides := map[string]interface{}{
			"chargeback_rate":  "not a number",
			"account_age_days": "also not a number",
		}

		service.applyOverrides(original, overrides)

		if original.ChargebackRate.InexactFloat64() != 0.5 {
			t.Error("invalid overrides should be ignored")
		}
		if original.AccountAgeDays != 100 {
			t.Error("invalid overrides should be ignored")
		}
	})
}
