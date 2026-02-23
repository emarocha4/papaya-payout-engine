package merchant

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Generator struct {
	rand *rand.Rand
}

func NewGenerator() *Generator {
	return &Generator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Generator) Generate(count int) []Merchant {
	merchants := make([]Merchant, 0, count)

	lowRiskCount := int(float64(count) * 0.60)
	mediumRiskCount := int(float64(count) * 0.25)
	highRiskCount := count - lowRiskCount - mediumRiskCount

	for i := 0; i < lowRiskCount; i++ {
		merchants = append(merchants, g.generateLowRiskMerchant())
	}
	for i := 0; i < mediumRiskCount; i++ {
		merchants = append(merchants, g.generateMediumRiskMerchant())
	}
	for i := 0; i < highRiskCount; i++ {
		merchants = append(merchants, g.generateHighRiskMerchant())
	}

	return merchants
}

func (g *Generator) generateLowRiskMerchant() Merchant {
	industries := []string{"RETAIL", "FOOD_DELIVERY", "SERVICES", "UTILITIES", "HEALTHCARE"}
	countries := []string{"BR", "MX", "AR", "CO", "CL", "PE", "UY"}

	industry := industries[g.rand.Intn(len(industries))]
	country := countries[g.rand.Intn(len(countries))]

	accountAgeDays := 365 + g.rand.Intn(1000)
	volume := 50000 + g.rand.Float64()*450000
	count := 200 + g.rand.Intn(2000)
	chargebackCount := g.rand.Intn(2)
	chargebackRate := float64(chargebackCount) / float64(count) * 100
	if chargebackRate > 0.5 {
		chargebackRate = g.rand.Float64() * 0.5
	}

	return Merchant{
		ID:                   uuid.New(),
		MerchantName:         g.generateMerchantName(industry),
		Industry:             industry,
		Country:              country,
		TransactionVolume30d: decimal.NewFromFloat(volume),
		TransactionCount30d:  count,
		AvgTicketSize:        decimal.NewFromFloat(volume / float64(count)),
		ChargebackCount30d:   chargebackCount,
		ChargebackRate:       decimal.NewFromFloat(chargebackRate),
		RefundRate:           decimal.NewFromFloat(1.0 + g.rand.Float64()*2.0),
		VelocityMultiplier:   decimal.NewFromFloat(0.8 + g.rand.Float64()*0.6),
		AccountAgeDays:       accountAgeDays,
		AccountCreatedAt:     time.Now().AddDate(0, 0, -accountAgeDays),
		KYCVerified:          true,
		KYCLevel:             g.randomChoice([]string{"FULL", "ENHANCED"}),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (g *Generator) generateMediumRiskMerchant() Merchant {
	industries := []string{"FASHION", "ELECTRONICS", "SERVICES", "RETAIL"}
	countries := []string{"BR", "MX", "AR", "CO", "CL"}

	industry := industries[g.rand.Intn(len(industries))]
	country := countries[g.rand.Intn(len(countries))]

	accountAgeDays := 90 + g.rand.Intn(275)
	volume := 20000 + g.rand.Float64()*80000
	count := 100 + g.rand.Intn(400)
	chargebackCount := 1 + g.rand.Intn(4)
	chargebackRate := float64(chargebackCount) / float64(count) * 100

	return Merchant{
		ID:                   uuid.New(),
		MerchantName:         g.generateMerchantName(industry),
		Industry:             industry,
		Country:              country,
		TransactionVolume30d: decimal.NewFromFloat(volume),
		TransactionCount30d:  count,
		AvgTicketSize:        decimal.NewFromFloat(volume / float64(count)),
		ChargebackCount30d:   chargebackCount,
		ChargebackRate:       decimal.NewFromFloat(chargebackRate),
		RefundRate:           decimal.NewFromFloat(3.0 + g.rand.Float64()*3.0),
		VelocityMultiplier:   decimal.NewFromFloat(1.5 + g.rand.Float64()*1.0),
		AccountAgeDays:       accountAgeDays,
		AccountCreatedAt:     time.Now().AddDate(0, 0, -accountAgeDays),
		KYCVerified:          true,
		KYCLevel:             "FULL",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (g *Generator) generateHighRiskMerchant() Merchant {
	industries := []string{"DIGITAL_GOODS", "TRAVEL", "ELECTRONICS"}
	countries := []string{"BR", "MX", "CO", "AR"}

	industry := industries[g.rand.Intn(len(industries))]
	country := countries[g.rand.Intn(len(countries))]

	var accountAgeDays int
	if g.rand.Float64() < 0.7 {
		accountAgeDays = 15 + g.rand.Intn(75)
	} else {
		accountAgeDays = 365 + g.rand.Intn(365)
	}

	volume := 5000 + g.rand.Float64()*545000
	count := 50 + g.rand.Intn(300)
	chargebackCount := 2 + g.rand.Intn(8)
	chargebackRate := float64(chargebackCount) / float64(count) * 100

	kycVerified := g.rand.Float64() > 0.4
	kycLevel := "NONE"
	if kycVerified {
		kycLevel = g.randomChoice([]string{"PARTIAL", "FULL"})
	}

	return Merchant{
		ID:                   uuid.New(),
		MerchantName:         g.generateMerchantName(industry),
		Industry:             industry,
		Country:              country,
		TransactionVolume30d: decimal.NewFromFloat(volume),
		TransactionCount30d:  count,
		AvgTicketSize:        decimal.NewFromFloat(volume / float64(count)),
		ChargebackCount30d:   chargebackCount,
		ChargebackRate:       decimal.NewFromFloat(chargebackRate),
		RefundRate:           decimal.NewFromFloat(6.0 + g.rand.Float64()*6.0),
		VelocityMultiplier:   decimal.NewFromFloat(2.5 + g.rand.Float64()*4.0),
		AccountAgeDays:       accountAgeDays,
		AccountCreatedAt:     time.Now().AddDate(0, 0, -accountAgeDays),
		KYCVerified:          kycVerified,
		KYCLevel:             kycLevel,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (g *Generator) generateMerchantName(industry string) string {
	prefixes := []string{"Global", "Premium", "Quick", "Express", "Pro", "Elite", "Smart", "Fast"}
	suffixes := map[string][]string{
		"RETAIL":        {"Store", "Shop", "Market", "Outlet"},
		"FOOD_DELIVERY": {"Eats", "Food", "Kitchen", "Delivery"},
		"SERVICES":      {"Services", "Solutions", "Partners", "Group"},
		"UTILITIES":     {"Utilities", "Energy", "Power", "Services"},
		"HEALTHCARE":    {"Health", "Medical", "Care", "Clinic"},
		"FASHION":       {"Fashion", "Apparel", "Boutique", "Style"},
		"ELECTRONICS":   {"Electronics", "Tech", "Digital", "Gadgets"},
		"DIGITAL_GOODS": {"Digital", "Media", "Downloads", "Content"},
		"TRAVEL":        {"Travel", "Tours", "Trips", "Adventures"},
	}

	prefix := prefixes[g.rand.Intn(len(prefixes))]
	suffix := suffixes[industry][g.rand.Intn(len(suffixes[industry]))]

	return prefix + " " + suffix
}

func (g *Generator) randomChoice(choices []string) string {
	return choices[g.rand.Intn(len(choices))]
}
