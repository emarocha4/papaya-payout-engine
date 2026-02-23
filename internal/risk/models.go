package risk

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RiskLevel string

const (
	RiskLevelLow        RiskLevel = "LOW"
	RiskLevelMediumLow  RiskLevel = "MEDIUM_LOW"
	RiskLevelMedium     RiskLevel = "MEDIUM"
	RiskLevelHigh       RiskLevel = "HIGH"
	RiskLevelCritical   RiskLevel = "CRITICAL"
)

type HoldPeriod string

const (
	HoldPeriodImmediate HoldPeriod = "IMMEDIATE"
	HoldPeriod7Days     HoldPeriod = "7_DAYS"
	HoldPeriod14Days    HoldPeriod = "14_DAYS"
	HoldPeriod45Days    HoldPeriod = "45_DAYS"
)

type RiskDecision struct {
	ID                       uuid.UUID        `json:"decision_id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID               uuid.UUID        `json:"merchant_id" gorm:"type:uuid;not null"`
	BatchID                  *uuid.UUID       `json:"batch_id,omitempty" gorm:"type:uuid"`
	RiskScore                int              `json:"risk_score" gorm:"not null"`
	RiskLevel                RiskLevel        `json:"risk_level" gorm:"not null"`
	PayoutHoldPeriod         HoldPeriod       `json:"payout_hold_period" gorm:"not null"`
	RollingReservePercentage int              `json:"rolling_reserve_percentage" gorm:"not null"`
	Reasoning                Reasoning        `json:"reasoning" gorm:"type:jsonb;not null"`
	EvaluatedAt              time.Time        `json:"evaluated_at" gorm:"not null;default:now()"`
	Simulation               bool             `json:"simulation" gorm:"not null;default:false"`
}

func (RiskDecision) TableName() string {
	return "risk_decisions"
}

type Reasoning struct {
	PrimaryFactors     []FactorExplanation `json:"primary_factors"`
	PolicyExplanation  string              `json:"policy_explanation"`
}

func (r *Reasoning) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, r)
}

func (r Reasoning) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type FactorExplanation struct {
	Factor       string `json:"factor"`
	Score        int    `json:"score"`
	Contribution string `json:"contribution"`
	Impact       string `json:"impact"`
}

type PolicyTier struct {
	MinScore                 int
	MaxScore                 int
	RiskLevel                RiskLevel
	HoldPeriod               HoldPeriod
	ReservePercentage        int
	Label                    string
}

type FactorScore struct {
	Chargeback     int
	AccountAge     int
	Velocity       int
	Category       int
	KYC            int
	Refund         int
}

type BatchReport struct {
	ID                 uuid.UUID              `json:"batch_id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TotalMerchants     int                    `json:"total_merchants" gorm:"not null"`
	Summary            BatchSummary           `json:"summary" gorm:"type:jsonb;not null"`
	HighRiskMerchants  []HighRiskMerchant     `json:"high_risk_merchants" gorm:"type:jsonb;not null"`
	Simulation         bool                   `json:"simulation" gorm:"not null;default:false"`
	CreatedAt          time.Time              `json:"evaluated_at" gorm:"not null;default:now()"`
}

func (BatchReport) TableName() string {
	return "batch_reports"
}

type BatchSummary struct {
	ByHoldPeriod   map[string]int     `json:"by_hold_period"`
	ByReserve      map[string]int     `json:"by_reserve"`
	ByRiskLevel    map[string]int     `json:"by_risk_level"`
	TotalVolume    float64            `json:"total_volume"`
	VolumeByTier   map[string]float64 `json:"volume_by_tier"`
}

func (bs *BatchSummary) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, bs)
}

func (bs BatchSummary) Value() (driver.Value, error) {
	return json.Marshal(bs)
}

type HighRiskMerchant struct {
	MerchantID         uuid.UUID `json:"merchant_id"`
	MerchantName       string    `json:"merchant_name"`
	RiskScore          int       `json:"risk_score"`
	RiskLevel          RiskLevel `json:"risk_level"`
	PrimaryConcerns    []string  `json:"primary_concerns"`
	RecommendedAction  string    `json:"recommended_action"`
}
