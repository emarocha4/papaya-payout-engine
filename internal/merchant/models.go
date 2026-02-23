package merchant

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Merchant struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantName string          `json:"merchant_name" gorm:"not null"`
	Industry     string          `json:"industry" gorm:"not null"`
	Country      string          `json:"country" gorm:"not null"`

	TransactionVolume30d decimal.Decimal `json:"transaction_volume_30d" gorm:"column:transaction_volume_30d;type:decimal(15,2);not null;default:0"`
	TransactionCount30d  int             `json:"transaction_count_30d" gorm:"column:transaction_count_30d;not null;default:0"`
	AvgTicketSize        decimal.Decimal `json:"avg_ticket_size" gorm:"column:avg_ticket_size;type:decimal(10,2);not null;default:0"`

	ChargebackCount30d  int             `json:"chargeback_count_30d" gorm:"column:chargeback_count_30d;not null;default:0"`
	ChargebackRate      decimal.Decimal `json:"chargeback_rate" gorm:"column:chargeback_rate;type:decimal(5,2);not null;default:0"`
	RefundRate          decimal.Decimal `json:"refund_rate" gorm:"column:refund_rate;type:decimal(5,2);not null;default:0"`
	VelocityMultiplier  decimal.Decimal `json:"velocity_multiplier" gorm:"column:velocity_multiplier;type:decimal(5,2);not null;default:1.0"`

	AccountAgeDays     int       `json:"account_age_days" gorm:"column:account_age_days;not null;default:0"`
	AccountCreatedAt   time.Time `json:"account_created_at" gorm:"column:account_created_at;not null;default:now()"`
	KYCVerified        bool      `json:"kyc_verified" gorm:"column:kyc_verified;not null;default:false"`
	KYCLevel           string    `json:"kyc_level" gorm:"column:kyc_level;not null;default:'NONE'"`

	CreatedAt time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

func (Merchant) TableName() string {
	return "merchants"
}

type MerchantProfile struct {
	MerchantID       uuid.UUID     `json:"merchant_id"`
	MerchantName     string        `json:"merchant_name"`
	Industry         string        `json:"industry"`
	Country          string        `json:"country"`
	AccountCreatedAt time.Time     `json:"account_created_at"`
	AccountAgeDays   int           `json:"account_age_days"`
	RiskMetrics      RiskMetrics   `json:"risk_metrics"`
	CurrentPolicy    *PolicyInfo   `json:"current_policy,omitempty"`
}

type RiskMetrics struct {
	TransactionVolume30d decimal.Decimal `json:"transaction_volume_30d"`
	TransactionCount30d  int             `json:"transaction_count_30d"`
	AvgTicketSize        decimal.Decimal `json:"avg_ticket_size"`
	ChargebackCount30d   int             `json:"chargeback_count_30d"`
	ChargebackRate       decimal.Decimal `json:"chargeback_rate"`
	RefundRate           decimal.Decimal `json:"refund_rate"`
	VelocityMultiplier   decimal.Decimal `json:"velocity_multiplier"`
	KYCVerified          bool            `json:"kyc_verified"`
	KYCLevel             string          `json:"kyc_level"`
}

type PolicyInfo struct {
	RiskScore                 int       `json:"risk_score"`
	PayoutHoldPeriod          string    `json:"payout_hold_period"`
	RollingReservePercentage  int       `json:"rolling_reserve_percentage"`
	LastEvaluatedAt           time.Time `json:"last_evaluated_at"`
}
