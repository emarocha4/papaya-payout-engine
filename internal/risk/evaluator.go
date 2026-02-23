package risk

import (
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
)

type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) CalculateChargebackScore(m *merchant.Merchant) int {
	rate := m.ChargebackRate.InexactFloat64()

	var baseScore int
	switch {
	case rate < 0.5:
		baseScore = 0
	case rate >= 0.5 && rate < 1.0:
		baseScore = 10
	case rate >= 1.0 && rate < 1.5:
		baseScore = 20
	default:
		baseScore = 30
	}

	return baseScore
}

func (e *Evaluator) CalculateAccountAgeScore(m *merchant.Merchant) int {
	days := m.AccountAgeDays

	switch {
	case days < 30:
		return 25
	case days >= 30 && days < 91:
		return 20
	case days >= 91 && days < 181:
		return 15
	case days >= 181 && days < 366:
		return 10
	case days >= 366 && days < 731:
		return 5
	default:
		return 0
	}
}

func (e *Evaluator) CalculateVelocityScore(m *merchant.Merchant) int {
	multiplier := m.VelocityMultiplier.InexactFloat64()

	switch {
	case multiplier < 1.5:
		return 0
	case multiplier >= 1.5 && multiplier < 2.5:
		return 5
	case multiplier >= 2.5 && multiplier < 4.0:
		return 10
	case multiplier >= 4.0 && multiplier < 6.0:
		return 15
	default:
		return 20
	}
}

func (e *Evaluator) CalculateCategoryScore(m *merchant.Merchant) int {
	switch m.Industry {
	case "DIGITAL_GOODS", "TRAVEL", "ELECTRONICS":
		return 15
	case "FASHION", "SERVICES":
		return 10
	case "FOOD_DELIVERY", "RETAIL":
		return 5
	case "UTILITIES", "HEALTHCARE":
		return 0
	default:
		return 10
	}
}

func (e *Evaluator) CalculateKYCScore(m *merchant.Merchant) int {
	if !m.KYCVerified {
		return 10
	}

	switch m.KYCLevel {
	case "NONE":
		return 10
	case "PARTIAL":
		return 7
	case "FULL":
		return 3
	case "ENHANCED":
		return 0
	default:
		return 10
	}
}

func (e *Evaluator) CalculateRefundScore(m *merchant.Merchant) int {
	rate := m.RefundRate.InexactFloat64()

	switch {
	case rate < 3.0:
		return 0
	case rate >= 3.0 && rate < 6.0:
		return 3
	default:
		return 5
	}
}

func (e *Evaluator) CalculateTotalScore(m *merchant.Merchant) (int, FactorScore) {
	factors := FactorScore{
		Chargeback: e.CalculateChargebackScore(m),
		AccountAge: e.CalculateAccountAgeScore(m),
		Velocity:   e.CalculateVelocityScore(m),
		Category:   e.CalculateCategoryScore(m),
		KYC:        e.CalculateKYCScore(m),
		Refund:     e.CalculateRefundScore(m),
	}

	total := factors.Chargeback + factors.AccountAge + factors.Velocity +
		factors.Category + factors.KYC + factors.Refund

	if total > 100 {
		total = 100
	}

	return total, factors
}
