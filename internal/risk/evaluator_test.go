package risk

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
)

func TestCalculateChargebackScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name string
		rate float64
		want int
	}{
		{"excellent - under 0.5%", 0.3, 0},
		{"acceptable - 0.5-1.0%", 0.8, 10},
		{"concerning - 1.0-1.5%", 1.2, 20},
		{"critical - over 1.5%", 4.5, 30},
		{"edge case - exactly 0.5%", 0.5, 10},
		{"edge case - exactly 1.0%", 1.0, 20},
		{"edge case - exactly 1.5%", 1.5, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &merchant.Merchant{
				ChargebackRate: decimal.NewFromFloat(tt.rate),
			}
			got := e.CalculateChargebackScore(m)
			if got != tt.want {
				t.Errorf("CalculateChargebackScore(%v) = %v, want %v", tt.rate, got, tt.want)
			}
		})
	}
}

func TestCalculateAccountAgeScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name string
		days int
		want int
	}{
		{"very new - under 30 days", 15, 25},
		{"new - 30-90 days", 60, 20},
		{"early stage - 91-180 days", 120, 15},
		{"established - 181-365 days", 300, 10},
		{"mature - 366-730 days", 500, 5},
		{"veteran - over 730 days", 800, 0},
		{"edge case - exactly 30 days", 30, 20},
		{"edge case - exactly 91 days", 91, 15},
		{"edge case - exactly 366 days", 366, 5},
		{"edge case - exactly 731 days", 731, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &merchant.Merchant{
				AccountAgeDays: tt.days,
			}
			got := e.CalculateAccountAgeScore(m)
			if got != tt.want {
				t.Errorf("CalculateAccountAgeScore(%v) = %v, want %v", tt.days, got, tt.want)
			}
		})
	}
}

func TestCalculateVelocityScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name       string
		multiplier float64
		want       int
	}{
		{"normal - under 1.5x", 1.2, 0},
		{"elevated - 1.5-2.5x", 2.0, 5},
		{"concerning - 2.5-4.0x", 3.0, 10},
		{"high risk - 4.0-6.0x", 5.0, 15},
		{"critical - over 6.0x", 8.0, 20},
		{"edge case - exactly 1.5x", 1.5, 5},
		{"edge case - exactly 4.0x", 4.0, 15},
		{"edge case - exactly 6.0x", 6.0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &merchant.Merchant{
				VelocityMultiplier: decimal.NewFromFloat(tt.multiplier),
			}
			got := e.CalculateVelocityScore(m)
			if got != tt.want {
				t.Errorf("CalculateVelocityScore(%v) = %v, want %v", tt.multiplier, got, tt.want)
			}
		})
	}
}

func TestCalculateCategoryScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name     string
		industry string
		want     int
	}{
		{"high risk - digital goods", "DIGITAL_GOODS", 15},
		{"high risk - travel", "TRAVEL", 15},
		{"high risk - electronics", "ELECTRONICS", 15},
		{"medium risk - fashion", "FASHION", 10},
		{"medium risk - services", "SERVICES", 10},
		{"low risk - food delivery", "FOOD_DELIVERY", 5},
		{"low risk - retail", "RETAIL", 5},
		{"minimal risk - utilities", "UTILITIES", 0},
		{"minimal risk - healthcare", "HEALTHCARE", 0},
		{"unknown category", "UNKNOWN", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &merchant.Merchant{
				Industry: tt.industry,
			}
			got := e.CalculateCategoryScore(m)
			if got != tt.want {
				t.Errorf("CalculateCategoryScore(%v) = %v, want %v", tt.industry, got, tt.want)
			}
		})
	}
}

func TestCalculateKYCScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name     string
		verified bool
		level    string
		want     int
	}{
		{"not verified", false, "NONE", 10},
		{"none level", true, "NONE", 10},
		{"partial level", true, "PARTIAL", 7},
		{"full level", true, "FULL", 3},
		{"enhanced level", true, "ENHANCED", 0},
		{"unknown level", true, "UNKNOWN", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &merchant.Merchant{
				KYCVerified: tt.verified,
				KYCLevel:    tt.level,
			}
			got := e.CalculateKYCScore(m)
			if got != tt.want {
				t.Errorf("CalculateKYCScore(verified=%v, level=%v) = %v, want %v", tt.verified, tt.level, got, tt.want)
			}
		})
	}
}

func TestCalculateRefundScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name string
		rate float64
		want int
	}{
		{"normal - under 3%", 2.5, 0},
		{"elevated - 3-6%", 4.0, 3},
		{"high - over 6%", 8.5, 5},
		{"edge case - exactly 3%", 3.0, 3},
		{"edge case - exactly 6%", 6.0, 5},
		{"very low", 0.5, 0},
		{"very high", 15.0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &merchant.Merchant{
				RefundRate: decimal.NewFromFloat(tt.rate),
			}
			got := e.CalculateRefundScore(m)
			if got != tt.want {
				t.Errorf("CalculateRefundScore(%v) = %v, want %v", tt.rate, got, tt.want)
			}
		})
	}
}

func TestCalculateTotalScore(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name      string
		merchant  *merchant.Merchant
		wantScore int
	}{
		{
			name: "low risk merchant",
			merchant: &merchant.Merchant{
				ChargebackRate:     decimal.NewFromFloat(0.3),
				AccountAgeDays:     800,
				VelocityMultiplier: decimal.NewFromFloat(1.2),
				Industry:           "RETAIL",
				KYCVerified:        true,
				KYCLevel:           "ENHANCED",
				RefundRate:         decimal.NewFromFloat(2.0),
			},
			wantScore: 5,
		},
		{
			name: "high risk merchant",
			merchant: &merchant.Merchant{
				ChargebackRate:     decimal.NewFromFloat(4.5),
				AccountAgeDays:     15,
				VelocityMultiplier: decimal.NewFromFloat(8.0),
				Industry:           "DIGITAL_GOODS",
				KYCVerified:        false,
				KYCLevel:           "NONE",
				RefundRate:         decimal.NewFromFloat(8.5),
			},
			wantScore: 100,
		},
		{
			name: "medium risk merchant",
			merchant: &merchant.Merchant{
				ChargebackRate:     decimal.NewFromFloat(0.8),
				AccountAgeDays:     120,
				VelocityMultiplier: decimal.NewFromFloat(2.0),
				Industry:           "FASHION",
				KYCVerified:        true,
				KYCLevel:           "FULL",
				RefundRate:         decimal.NewFromFloat(4.0),
			},
			wantScore: 46,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, factors := e.CalculateTotalScore(tt.merchant)

			if score != tt.wantScore {
				t.Errorf("CalculateTotalScore() score = %v, want %v", score, tt.wantScore)
				t.Logf("Factor breakdown: Chargeback=%d, AccountAge=%d, Velocity=%d, Category=%d, KYC=%d, Refund=%d",
					factors.Chargeback, factors.AccountAge, factors.Velocity, factors.Category, factors.KYC, factors.Refund)
			}

			if factors.Chargeback < 0 || factors.Chargeback > 30 {
				t.Errorf("Chargeback score %d out of valid range [0, 30]", factors.Chargeback)
			}
			if factors.AccountAge < 0 || factors.AccountAge > 25 {
				t.Errorf("AccountAge score %d out of valid range [0, 25]", factors.AccountAge)
			}
			if factors.Velocity < 0 || factors.Velocity > 20 {
				t.Errorf("Velocity score %d out of valid range [0, 20]", factors.Velocity)
			}
			if factors.Category < 0 || factors.Category > 15 {
				t.Errorf("Category score %d out of valid range [0, 15]", factors.Category)
			}
			if factors.KYC < 0 || factors.KYC > 10 {
				t.Errorf("KYC score %d out of valid range [0, 10]", factors.KYC)
			}
			if factors.Refund < 0 || factors.Refund > 5 {
				t.Errorf("Refund score %d out of valid range [0, 5]", factors.Refund)
			}
		})
	}
}
