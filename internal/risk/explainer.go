package risk

import (
	"fmt"

	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
)

type Explainer struct{}

func NewExplainer() *Explainer {
	return &Explainer{}
}

func (e *Explainer) GenerateReasoning(m *merchant.Merchant, factors FactorScore, tier PolicyTier) Reasoning {
	primaryFactors := []FactorExplanation{
		e.ExplainChargebackScore(factors.Chargeback, m.ChargebackRate.InexactFloat64()),
		e.ExplainAccountAgeScore(factors.AccountAge, m.AccountAgeDays),
		e.ExplainVelocityScore(factors.Velocity, m.VelocityMultiplier.InexactFloat64()),
		e.ExplainCategoryScore(factors.Category, m.Industry),
		e.ExplainKYCScore(factors.KYC, m.KYCVerified, m.KYCLevel),
		e.ExplainRefundScore(factors.Refund, m.RefundRate.InexactFloat64()),
	}

	totalScore := factors.Chargeback + factors.AccountAge + factors.Velocity +
		factors.Category + factors.KYC + factors.Refund

	policyExplanation := fmt.Sprintf(
		"Score of %d places merchant in %s tier requiring %s hold and %d%% reserve",
		totalScore, tier.RiskLevel, tier.HoldPeriod, tier.ReservePercentage,
	)

	return Reasoning{
		PrimaryFactors:    primaryFactors,
		PolicyExplanation: policyExplanation,
	}
}

func (e *Explainer) ExplainChargebackScore(score int, rate float64) FactorExplanation {
	var contribution string
	var impact string

	switch {
	case rate < 0.5:
		contribution = fmt.Sprintf("%.2f%% rate - Excellent", rate)
		impact = "POSITIVE"
	case rate >= 0.5 && rate < 1.0:
		contribution = fmt.Sprintf("%.2f%% rate - Acceptable range", rate)
		impact = "NEUTRAL"
	case rate >= 1.0 && rate < 1.5:
		contribution = fmt.Sprintf("%.2f%% rate - Concerning", rate)
		impact = "NEGATIVE"
	default:
		contribution = fmt.Sprintf("%.2f%% rate - Critical", rate)
		impact = "CRITICAL"
	}

	return FactorExplanation{
		Factor:       "Chargeback Rate",
		Score:        score,
		Contribution: contribution,
		Impact:       impact,
	}
}

func (e *Explainer) ExplainAccountAgeScore(score int, days int) FactorExplanation {
	var contribution string
	var impact string

	switch {
	case days < 30:
		contribution = fmt.Sprintf("Account %d days old - Very new", days)
		impact = "CRITICAL"
	case days >= 30 && days < 91:
		contribution = fmt.Sprintf("Account %d days old - New", days)
		impact = "NEGATIVE"
	case days >= 91 && days < 181:
		contribution = fmt.Sprintf("Account %d days old - Early stage", days)
		impact = "NEUTRAL"
	case days >= 181 && days < 366:
		contribution = fmt.Sprintf("Account %d days old - Established", days)
		impact = "NEUTRAL"
	case days >= 366 && days < 731:
		contribution = fmt.Sprintf("Account %d days old - Mature", days)
		impact = "POSITIVE"
	default:
		contribution = fmt.Sprintf("Account %d days old - Veteran", days)
		impact = "POSITIVE"
	}

	return FactorExplanation{
		Factor:       "Account Age",
		Score:        score,
		Contribution: contribution,
		Impact:       impact,
	}
}

func (e *Explainer) ExplainVelocityScore(score int, multiplier float64) FactorExplanation {
	var contribution string
	var impact string

	switch {
	case multiplier < 1.5:
		contribution = fmt.Sprintf("%.1fx velocity - Normal", multiplier)
		impact = "POSITIVE"
	case multiplier >= 1.5 && multiplier < 2.5:
		contribution = fmt.Sprintf("%.1fx velocity - Elevated", multiplier)
		impact = "NEUTRAL"
	case multiplier >= 2.5 && multiplier < 4.0:
		contribution = fmt.Sprintf("%.1fx velocity - Concerning", multiplier)
		impact = "NEGATIVE"
	case multiplier >= 4.0 && multiplier < 6.0:
		contribution = fmt.Sprintf("%.1fx velocity - High risk", multiplier)
		impact = "NEGATIVE"
	default:
		contribution = fmt.Sprintf("%.1fx velocity - Critical", multiplier)
		impact = "CRITICAL"
	}

	return FactorExplanation{
		Factor:       "Transaction Velocity",
		Score:        score,
		Contribution: contribution,
		Impact:       impact,
	}
}

func (e *Explainer) ExplainCategoryScore(score int, industry string) FactorExplanation {
	var contribution string
	var impact string

	switch score {
	case 15:
		contribution = fmt.Sprintf("%s - High risk category", industry)
		impact = "NEGATIVE"
	case 10:
		contribution = fmt.Sprintf("%s - Medium risk category", industry)
		impact = "NEUTRAL"
	case 5:
		contribution = fmt.Sprintf("%s - Low risk category", industry)
		impact = "POSITIVE"
	case 0:
		contribution = fmt.Sprintf("%s - Minimal risk category", industry)
		impact = "POSITIVE"
	default:
		contribution = fmt.Sprintf("%s - Unknown category", industry)
		impact = "NEUTRAL"
	}

	return FactorExplanation{
		Factor:       "Business Category",
		Score:        score,
		Contribution: contribution,
		Impact:       impact,
	}
}

func (e *Explainer) ExplainKYCScore(score int, verified bool, level string) FactorExplanation {
	var contribution string
	var impact string

	if !verified {
		contribution = "No KYC verification"
		impact = "CRITICAL"
	} else {
		switch level {
		case "ENHANCED":
			contribution = "Enhanced KYC - Full business documentation"
			impact = "POSITIVE"
		case "FULL":
			contribution = "Full KYC - ID and address verified"
			impact = "NEUTRAL"
		case "PARTIAL":
			contribution = "Partial KYC - ID only"
			impact = "NEGATIVE"
		default:
			contribution = "No KYC verification"
			impact = "CRITICAL"
		}
	}

	return FactorExplanation{
		Factor:       "KYC Verification",
		Score:        score,
		Contribution: contribution,
		Impact:       impact,
	}
}

func (e *Explainer) ExplainRefundScore(score int, rate float64) FactorExplanation {
	var contribution string
	var impact string

	switch {
	case rate < 3.0:
		contribution = fmt.Sprintf("%.1f%% refund rate - Normal", rate)
		impact = "POSITIVE"
	case rate >= 3.0 && rate < 6.0:
		contribution = fmt.Sprintf("%.1f%% refund rate - Elevated", rate)
		impact = "NEUTRAL"
	default:
		contribution = fmt.Sprintf("%.1f%% refund rate - High (fraud signal)", rate)
		impact = "NEGATIVE"
	}

	return FactorExplanation{
		Factor:       "Refund Rate",
		Score:        score,
		Contribution: contribution,
		Impact:       impact,
	}
}
