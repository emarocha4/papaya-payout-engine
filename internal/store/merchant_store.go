package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
	"gorm.io/gorm"
)

type MerchantStore struct {
	db *gorm.DB
}

func NewMerchantStore(db *gorm.DB) *MerchantStore {
	return &MerchantStore{db: db}
}

func (s *MerchantStore) Create(ctx context.Context, m *merchant.Merchant) error {
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("failed to create merchant: %w", err)
	}
	return nil
}

func (s *MerchantStore) Get(ctx context.Context, id uuid.UUID) (*merchant.Merchant, error) {
	var m merchant.Merchant
	if err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("merchant not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}
	return &m, nil
}

func (s *MerchantStore) List(ctx context.Context, limit, offset int) ([]merchant.Merchant, int64, error) {
	var merchants []merchant.Merchant
	var total int64

	if err := s.db.WithContext(ctx).Model(&merchant.Merchant{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count merchants: %w", err)
	}

	if err := s.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&merchants).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list merchants: %w", err)
	}

	return merchants, total, nil
}

func (s *MerchantStore) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := s.db.WithContext(ctx).
		Model(&merchant.Merchant{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}
	return nil
}

func (s *MerchantStore) BulkCreate(ctx context.Context, merchants []merchant.Merchant) error {
	if err := s.db.WithContext(ctx).Create(&merchants).Error; err != nil {
		return fmt.Errorf("failed to bulk create merchants: %w", err)
	}
	return nil
}
