package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
)

type MerchantRepository interface {
	Get(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error)
}

type DecisionRepository interface {
	Create(ctx context.Context, decision *RiskDecision) error
	GetLatestByMerchant(ctx context.Context, merchantID uuid.UUID) (*RiskDecision, error)
	BulkCreate(ctx context.Context, decisions []RiskDecision) error
}

type Service struct {
	merchantStore MerchantRepository
	decisionStore DecisionRepository
	evaluator     *Evaluator
	policy        *PolicyMapper
	explainer     *Explainer
}

func NewService(
	merchantStore MerchantRepository,
	decisionStore DecisionRepository,
) *Service {
	return &Service{
		merchantStore: merchantStore,
		decisionStore: decisionStore,
		evaluator:     NewEvaluator(),
		policy:        NewPolicyMapper(),
		explainer:     NewExplainer(),
	}
}

func (s *Service) EvaluateMerchant(ctx context.Context, merchantID uuid.UUID, simulation bool) (*RiskDecision, error) {
	m, err := s.merchantStore.Get(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	totalScore, factors := s.evaluator.CalculateTotalScore(m)
	tier := s.policy.DeterminePolicyTier(totalScore)
	reasoning := s.explainer.GenerateReasoning(m, factors, tier)

	decision := &RiskDecision{
		MerchantID:               merchantID,
		RiskScore:                totalScore,
		RiskLevel:                tier.RiskLevel,
		PayoutHoldPeriod:         tier.HoldPeriod,
		RollingReservePercentage: tier.ReservePercentage,
		Reasoning:                reasoning,
		EvaluatedAt:              time.Now(),
		Simulation:               simulation,
	}

	if !simulation {
		if err := s.decisionStore.Create(ctx, decision); err != nil {
			return nil, fmt.Errorf("failed to save decision: %w", err)
		}
	}

	return decision, nil
}

func (s *Service) SimulateMerchant(ctx context.Context, merchantID uuid.UUID, overrides map[string]interface{}) (*RiskDecision, error) {
	m, err := s.merchantStore.Get(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	simulatedMerchant := *m
	s.applyOverrides(&simulatedMerchant, overrides)

	totalScore, factors := s.evaluator.CalculateTotalScore(&simulatedMerchant)
	tier := s.policy.DeterminePolicyTier(totalScore)
	reasoning := s.explainer.GenerateReasoning(&simulatedMerchant, factors, tier)

	decision := &RiskDecision{
		MerchantID:               merchantID,
		RiskScore:                totalScore,
		RiskLevel:                tier.RiskLevel,
		PayoutHoldPeriod:         tier.HoldPeriod,
		RollingReservePercentage: tier.ReservePercentage,
		Reasoning:                reasoning,
		EvaluatedAt:              time.Now(),
		Simulation:               true,
	}

	return decision, nil
}

func (s *Service) GetMerchantProfile(ctx context.Context, merchantID uuid.UUID) (*merchant.MerchantProfile, error) {
	m, err := s.merchantStore.Get(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	latestDecision, err := s.decisionStore.GetLatestByMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest decision: %w", err)
	}

	profile := &merchant.MerchantProfile{
		MerchantID:       m.ID,
		MerchantName:     m.MerchantName,
		Industry:         m.Industry,
		Country:          m.Country,
		AccountCreatedAt: m.AccountCreatedAt,
		AccountAgeDays:   m.AccountAgeDays,
		RiskMetrics: merchant.RiskMetrics{
			TransactionVolume30d: m.TransactionVolume30d,
			TransactionCount30d:  m.TransactionCount30d,
			AvgTicketSize:        m.AvgTicketSize,
			ChargebackCount30d:   m.ChargebackCount30d,
			ChargebackRate:       m.ChargebackRate,
			RefundRate:           m.RefundRate,
			VelocityMultiplier:   m.VelocityMultiplier,
			KYCVerified:          m.KYCVerified,
			KYCLevel:             m.KYCLevel,
		},
	}

	if latestDecision != nil {
		profile.CurrentPolicy = &merchant.PolicyInfo{
			RiskScore:                latestDecision.RiskScore,
			PayoutHoldPeriod:         string(latestDecision.PayoutHoldPeriod),
			RollingReservePercentage: latestDecision.RollingReservePercentage,
			LastEvaluatedAt:          latestDecision.EvaluatedAt,
		}
	}

	return profile, nil
}

func (s *Service) applyOverrides(m *merchant.Merchant, overrides map[string]interface{}) {
	if val, ok := overrides["chargeback_rate"].(float64); ok {
		m.ChargebackRate = decimal.NewFromFloat(val)
	}
	if val, ok := overrides["account_age_days"].(float64); ok {
		m.AccountAgeDays = int(val)
	}
	if val, ok := overrides["kyc_verified"].(bool); ok {
		m.KYCVerified = val
		if val && m.KYCLevel == "NONE" {
			m.KYCLevel = "FULL"
		}
	}
	if val, ok := overrides["velocity_multiplier"].(float64); ok {
		m.VelocityMultiplier = decimal.NewFromFloat(val)
	}
}
