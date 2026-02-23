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
