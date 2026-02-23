package risk

import "testing"

func TestDeterminePolicyTier(t *testing.T) {
	p := NewPolicyMapper()

	tests := []struct {
		name        string
		score       int
		wantLevel   RiskLevel
		wantHold    HoldPeriod
		wantReserve int
	}{
		{
			name:        "excellent - score 0",
			score:       0,
			wantLevel:   RiskLevelLow,
			wantHold:    HoldPeriodImmediate,
			wantReserve: 0,
		},
		{
			name:        "low risk - score 10",
			score:       10,
			wantLevel:   RiskLevelLow,
			wantHold:    HoldPeriodImmediate,
			wantReserve: 0,
		},
		{
			name:        "low risk - score 20",
			score:       20,
			wantLevel:   RiskLevelLow,
			wantHold:    HoldPeriodImmediate,
			wantReserve: 0,
		},
		{
			name:        "medium-low - score 21",
			score:       21,
			wantLevel:   RiskLevelMediumLow,
			wantHold:    HoldPeriod7Days,
			wantReserve: 0,
		},
		{
			name:        "medium-low - score 33",
			score:       33,
			wantLevel:   RiskLevelMediumLow,
			wantHold:    HoldPeriod7Days,
			wantReserve: 0,
		},
		{
			name:        "medium-low - score 40",
			score:       40,
			wantLevel:   RiskLevelMediumLow,
			wantHold:    HoldPeriod7Days,
			wantReserve: 0,
		},
		{
			name:        "medium - score 50",
			score:       50,
			wantLevel:   RiskLevelMedium,
			wantHold:    HoldPeriod14Days,
			wantReserve: 10,
		},
		{
			name:        "medium - score 60",
			score:       60,
			wantLevel:   RiskLevelMedium,
			wantHold:    HoldPeriod14Days,
			wantReserve: 10,
		},
		{
			name:        "high - score 75",
			score:       75,
			wantLevel:   RiskLevelHigh,
			wantHold:    HoldPeriod45Days,
			wantReserve: 20,
		},
		{
			name:        "critical - score 81",
			score:       81,
			wantLevel:   RiskLevelCritical,
			wantHold:    HoldPeriod45Days,
			wantReserve: 20,
		},
		{
			name:        "critical - score 90",
			score:       90,
			wantLevel:   RiskLevelCritical,
			wantHold:    HoldPeriod45Days,
			wantReserve: 20,
		},
		{
			name:        "critical - score 100",
			score:       100,
			wantLevel:   RiskLevelCritical,
			wantHold:    HoldPeriod45Days,
			wantReserve: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := p.DeterminePolicyTier(tt.score)

			if tier.RiskLevel != tt.wantLevel {
				t.Errorf("score %d: got level %v, want %v", tt.score, tier.RiskLevel, tt.wantLevel)
			}
			if tier.HoldPeriod != tt.wantHold {
				t.Errorf("score %d: got hold %v, want %v", tt.score, tier.HoldPeriod, tt.wantHold)
			}
			if tier.ReservePercentage != tt.wantReserve {
				t.Errorf("score %d: got reserve %v%%, want %v%%", tt.score, tier.ReservePercentage, tt.wantReserve)
			}

			if tier.MinScore > tt.score || tier.MaxScore < tt.score {
				t.Errorf("score %d not within tier range [%d, %d]", tt.score, tier.MinScore, tier.MaxScore)
			}
		})
	}
}

func TestPolicyTierBoundaries(t *testing.T) {
	p := NewPolicyMapper()

	t.Run("all tiers are accessible", func(t *testing.T) {
		levels := []RiskLevel{
			RiskLevelLow,
			RiskLevelMediumLow,
			RiskLevelMedium,
			RiskLevelHigh,
			RiskLevelCritical,
		}

		foundLevels := make(map[RiskLevel]bool)

		for score := 0; score <= 100; score++ {
			tier := p.DeterminePolicyTier(score)
			foundLevels[tier.RiskLevel] = true
		}

		for _, level := range levels {
			if !foundLevels[level] {
				t.Errorf("Risk level %v is not reachable by any score", level)
			}
		}
	})

	t.Run("no gaps in score ranges", func(t *testing.T) {
		for score := 0; score < 100; score++ {
			tier1 := p.DeterminePolicyTier(score)
			tier2 := p.DeterminePolicyTier(score + 1)

			if tier1.RiskLevel != tier2.RiskLevel {
				if tier1.MaxScore != score {
					t.Errorf("gap detected: score %d is in tier ending at %d", score, tier1.MaxScore)
				}
				if tier2.MinScore != score+1 {
					t.Errorf("gap detected: score %d is in tier starting at %d", score+1, tier2.MinScore)
				}
			}
		}
	})

	t.Run("reserve increases with risk", func(t *testing.T) {
		lowTier := p.DeterminePolicyTier(10)
		mediumTier := p.DeterminePolicyTier(50)
		highTier := p.DeterminePolicyTier(75)

		if lowTier.ReservePercentage > mediumTier.ReservePercentage {
			t.Error("low tier reserve should not exceed medium tier reserve")
		}
		if mediumTier.ReservePercentage > highTier.ReservePercentage {
			t.Error("medium tier reserve should not exceed high tier reserve")
		}
	})
}
