package sale

import (
	"context"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSale(ctx context.Context, input CreateSaleInput) (*Sales, error) {
	ts := input.TimeStamp
	if ts.IsZero() {
		ts = time.Now()
	}

	sale := &Sales{
		Element:          input.Element,
		Price:            input.Price,
		TimeStamp:        ts,
		PayementMethodID: input.PayementMethodID,
	}
	return s.repo.Create(ctx, sale)
}

func (s *Service) GetSaleByID(ctx context.Context, id int) (*Sales, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetAllSales(ctx context.Context) ([]*Sales, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) UpdateSale(ctx context.Context, id int, input UpdateSaleInput) (*Sales, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Element != nil {
		existing.Element = *input.Element
	}
	if input.Price != nil {
		existing.Price = *input.Price
	}
	if input.TimeStamp != nil {
		existing.TimeStamp = *input.TimeStamp
	}
	if input.PayementMethodID != nil {
		existing.PayementMethodID = *input.PayementMethodID
	}

	return s.repo.Update(ctx, existing)
}

func (s *Service) DeleteSale(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
