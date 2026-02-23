package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
	"gorm.io/gorm"
)

type BatchStore struct {
	db *gorm.DB
}

func NewBatchStore(db *gorm.DB) *BatchStore {
	return &BatchStore{db: db}
}

func (s *BatchStore) Create(ctx context.Context, batch *risk.BatchReport) error {
	if err := s.db.WithContext(ctx).Create(batch).Error; err != nil {
		return fmt.Errorf("failed to create batch report: %w", err)
	}
	return nil
}

func (s *BatchStore) Get(ctx context.Context, id uuid.UUID) (*risk.BatchReport, error) {
	var batch risk.BatchReport
	if err := s.db.WithContext(ctx).First(&batch, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("batch report not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get batch report: %w", err)
	}
	return &batch, nil
}
