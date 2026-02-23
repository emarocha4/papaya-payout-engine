package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
	"gorm.io/gorm"
)

type DecisionStore struct {
	db *gorm.DB
}

func NewDecisionStore(db *gorm.DB) *DecisionStore {
	return &DecisionStore{db: db}
}

func (s *DecisionStore) Create(ctx context.Context, decision *risk.RiskDecision) error {
	if err := s.db.WithContext(ctx).Create(decision).Error; err != nil {
		return fmt.Errorf("failed to create decision: %w", err)
	}
	return nil
}

func (s *DecisionStore) GetLatestByMerchant(ctx context.Context, merchantID uuid.UUID) (*risk.RiskDecision, error) {
	var decision risk.RiskDecision
	if err := s.db.WithContext(ctx).
		Where("merchant_id = ? AND simulation = false", merchantID).
		Order("evaluated_at DESC").
		First(&decision).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest decision: %w", err)
	}
	return &decision, nil
}

func (s *DecisionStore) ListByBatch(ctx context.Context, batchID uuid.UUID) ([]risk.RiskDecision, error) {
	var decisions []risk.RiskDecision
	if err := s.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Find(&decisions).Error; err != nil {
		return nil, fmt.Errorf("failed to list decisions by batch: %w", err)
	}
	return decisions, nil
}

func (s *DecisionStore) BulkCreate(ctx context.Context, decisions []risk.RiskDecision) error {
	if err := s.db.WithContext(ctx).Create(&decisions).Error; err != nil {
		return fmt.Errorf("failed to bulk create decisions: %w", err)
	}
	return nil
}
