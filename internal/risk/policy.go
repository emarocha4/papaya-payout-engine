package risk

type PolicyMapper struct {
	tiers []PolicyTier
}

func NewPolicyMapper() *PolicyMapper {
	return &PolicyMapper{
		tiers: []PolicyTier{
			{
				MinScore:          0,
				MaxScore:          20,
				RiskLevel:         RiskLevelLow,
				HoldPeriod:        HoldPeriodImmediate,
				ReservePercentage: 0,
				Label:             "Low Risk - Trusted Merchant",
			},
			{
				MinScore:          21,
				MaxScore:          40,
				RiskLevel:         RiskLevelMediumLow,
				HoldPeriod:        HoldPeriod7Days,
				ReservePercentage: 0,
				Label:             "Medium-Low Risk - Standard Processing",
			},
			{
				MinScore:          41,
				MaxScore:          60,
				RiskLevel:         RiskLevelMedium,
				HoldPeriod:        HoldPeriod14Days,
				ReservePercentage: 10,
				Label:             "Medium Risk - Enhanced Monitoring",
			},
			{
				MinScore:          61,
				MaxScore:          80,
				RiskLevel:         RiskLevelHigh,
				HoldPeriod:        HoldPeriod45Days,
				ReservePercentage: 20,
				Label:             "High Risk - Requires Review",
			},
			{
				MinScore:          81,
				MaxScore:          100,
				RiskLevel:         RiskLevelCritical,
				HoldPeriod:        HoldPeriod45Days,
				ReservePercentage: 20,
				Label:             "Critical Risk - Manual Approval Required",
			},
		},
	}
}

func (p *PolicyMapper) DeterminePolicyTier(score int) PolicyTier {
	for _, tier := range p.tiers {
		if score >= tier.MinScore && score <= tier.MaxScore {
			return tier
		}
	}
	return p.tiers[len(p.tiers)-1]
}

func (p *PolicyMapper) GetHoldPeriod(score int) HoldPeriod {
	return p.DeterminePolicyTier(score).HoldPeriod
}

func (p *PolicyMapper) GetReservePercentage(score int) int {
	return p.DeterminePolicyTier(score).ReservePercentage
}
