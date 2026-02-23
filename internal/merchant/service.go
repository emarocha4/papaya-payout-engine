package merchant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type MerchantRepository interface {
	Create(ctx context.Context, m *Merchant) error
	Get(ctx context.Context, id uuid.UUID) (*Merchant, error)
	List(ctx context.Context, limit, offset int) ([]Merchant, int64, error)
	BulkCreate(ctx context.Context, merchants []Merchant) error
}

type Service struct {
	store MerchantRepository
}

func NewService(store MerchantRepository) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, m *Merchant) (*Merchant, error) {
	if err := s.store.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to create merchant: %w", err)
	}
	return m, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Merchant, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Merchant, int64, error) {
	return s.store.List(ctx, limit, offset)
}

func (s *Service) Seed(ctx context.Context, count int) ([]Merchant, error) {
	generator := NewGenerator()
	merchants := generator.Generate(count)

	if err := s.store.BulkCreate(ctx, merchants); err != nil {
		return nil, fmt.Errorf("failed to seed merchants: %w", err)
	}

	return merchants, nil
}
