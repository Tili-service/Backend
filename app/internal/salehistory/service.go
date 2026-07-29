package salehistory

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListBySaleID(ctx context.Context, saleID uuid.UUID) ([]*SaleHistory, error) {
	return s.repo.ListBySaleID(ctx, saleID)
}
