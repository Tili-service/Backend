package store

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateStoreInput) (*Store, error) {
	store := &Store{
		Name:      input.Name,
		BuyerID:   input.BuyerID,
		LicenceID: input.LicenceID,
		NumeroTVA: input.NumeroTVA,
		Siret:     input.Siret,
	}
	return s.repo.Create(ctx, store)
}

func (s *Service) FindByID(ctx context.Context, id int64) (*Store, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) FindByBuyerID(ctx context.Context, buyerID int64) ([]Store, error) {
	return s.repo.FindByBuyerID(ctx, buyerID)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}


