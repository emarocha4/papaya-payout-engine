package constants

import "time"

const (
	MaxBatchSize      = 100
	DefaultBatchSize  = 20
	DefaultWorkers    = 10
	BatchTimeout      = 30 * time.Second
	ServiceTimeout    = 10 * time.Second
	DefaultQueryLimit = 20
	MaxQueryLimit     = 100
)

const (
	DefaultChargebackExcellent   = 0.5
	DefaultChargebackAcceptable  = 1.0
	DefaultChargebackCritical    = 1.5
	DefaultVelocityNormal        = 1.5
	DefaultVelocityElevated      = 2.5
	DefaultVelocityConcerning    = 4.0
	DefaultVelocityHighRisk      = 6.0
	DefaultRefundNormal          = 3.0
	DefaultRefundElevated        = 6.0
)

const (
	AccountAgeVeryNew     = 30
	AccountAgeNew         = 91
	AccountAgeEarly       = 181
	AccountAgeEstablished = 366
	AccountAgeMature      = 731
)
