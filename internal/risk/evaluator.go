package risk

import (
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
	"github.com/yuno-payments/papaya-payout-engine/internal/platform/constants"
)

type Evaluator struct {
	chargebackExcellent  float64
	chargebackAcceptable float64
	chargebackCritical   float64
	velocityNormal       float64
	velocityElevated     float64
	velocityConcerning   float64
	velocityHighRisk     float64
	refundNormal         float64
	refundElevated       float64
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		chargebackExcellent:  constants.DefaultChargebackExcellent,
		chargebackAcceptable: constants.DefaultChargebackAcceptable,
		chargebackCritical:   constants.DefaultChargebackCritical,
		velocityNormal:       constants.DefaultVelocityNormal,
		velocityElevated:     constants.DefaultVelocityElevated,
		velocityConcerning:   constants.DefaultVelocityConcerning,
		velocityHighRisk:     constants.DefaultVelocityHighRisk,
		refundNormal:         constants.DefaultRefundNormal,
		refundElevated:       constants.DefaultRefundElevated,
	}
}

func NewEvaluatorWithThresholds(thresholds map[string]interface{}) *Evaluator {
	e := NewEvaluator()

	if val, ok := thresholds["chargeback_excellent"].(float64); ok {
		e.chargebackExcellent = val
	}
	if val, ok := thresholds["chargeback_acceptable"].(float64); ok {
		e.chargebackAcceptable = val
	}
	if val, ok := thresholds["chargeback_critical"].(float64); ok {
		e.chargebackCritical = val
	}
	if val, ok := thresholds["velocity_normal"].(float64); ok {
		e.velocityNormal = val
	}
	if val, ok := thresholds["refund_normal"].(float64); ok {
		e.refundNormal = val
	}
	if val, ok := thresholds["refund_elevated"].(float64); ok {
		e.refundElevated = val
	}

	return e
}

// CalculateChargebackScore evaluates the merchant's chargeback rate and assigns
// a risk score from 0-30 points. Higher chargeback rates indicate potential fraud,
// operational issues, or customer dissatisfaction.
//
// Scoring thresholds (configurable):
//   - < 0.5%: 0 points (excellent, industry best practice)
//   - 0.5-1.0%: 10 points (acceptable range)
//   - 1.0-1.5%: 20 points (concerning, approaching processor limits)
//   - > 1.5%: 30 points (critical, exceeds most processor thresholds)
//
// Note: Most payment processors enforce 1.5% maximum chargeback rate.
func (e *Evaluator) CalculateChargebackScore(m *merchant.Merchant) int {
	rate := m.ChargebackRate.InexactFloat64()

	var baseScore int
	switch {
	case rate < e.chargebackExcellent:
		baseScore = 0
	case rate >= e.chargebackExcellent && rate < e.chargebackAcceptable:
		baseScore = 10
	case rate >= e.chargebackAcceptable && rate < e.chargebackCritical:
		baseScore = 20
	default:
		baseScore = 30
	}

	return baseScore
}

// CalculateAccountAgeScore evaluates merchant account maturity and assigns
// a risk score from 0-25 points. Newer merchants have less established track records
// and higher risk profiles.
//
// Scoring thresholds:
//   - < 30 days: 25 points (very new, high-risk window)
//   - 30-90 days: 20 points (new, still establishing patterns)
//   - 91-180 days: 15 points (early stage)
//   - 181-365 days: 10 points (established)
//   - 366-730 days: 5 points (mature)
//   - > 730 days: 0 points (veteran, proven track record)
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
	case multiplier < e.velocityNormal:
		return 0
	case multiplier >= e.velocityNormal && multiplier < e.velocityElevated:
		return 5
	case multiplier >= e.velocityElevated && multiplier < e.velocityConcerning:
		return 10
	case multiplier >= e.velocityConcerning && multiplier < e.velocityHighRisk:
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

// CalculateRefundScore evaluates the merchant's refund rate and assigns
// a risk score from 0-5 points. Elevated refund rates may indicate fraud patterns,
// poor product quality, or misleading product descriptions.
//
// Scoring thresholds (configurable):
//   - < 3.0%: 0 points (normal operations)
//   - 3.0-6.0%: 3 points (elevated but acceptable)
//   - > 6.0%: 5 points (high risk, potential fraud signal)
//
// High refund rates combined with low chargebacks may indicate friendly fraud
// or merchant refunding to avoid chargebacks.
func (e *Evaluator) CalculateRefundScore(m *merchant.Merchant) int {
	rate := m.RefundRate.InexactFloat64()

	switch {
	case rate < e.refundNormal:
		return 0
	case rate >= e.refundNormal && rate < e.refundElevated:
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
