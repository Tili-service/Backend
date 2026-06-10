package salehistory

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListBySaleID(ctx context.Context, saleID int) ([]*SaleHistory, error) {
	return s.repo.ListBySaleID(ctx, saleID)
}
