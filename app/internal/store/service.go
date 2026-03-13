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

func (s *Service) Create(ctx context.Context, input CreateStoreInput, accountID int) (*Store, error) {
	store := &Store{
		Name:      input.Name,
		BuyerID:   accountID,
		LicenceID: input.LicenceID,
		NumeroTVA: input.NumeroTVA,
		Siret:     input.Siret,
	}
	return s.repo.Create(ctx, store)
}

func (s *Service) FindByID(ctx context.Context, id int) (*Store, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) FindByBuyerID(ctx context.Context, buyerID int) ([]Store, error) {
	return s.repo.FindByBuyerID(ctx, buyerID)
}

func (s *Service) FindAll(ctx context.Context) ([]*Store, error) {
	stores, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return stores, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, input UpdateStoreInput) (*Store, error) {
	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		store.Name = *input.Name
	}
	if input.NumeroTVA != nil {
		store.NumeroTVA = *input.NumeroTVA
	}
	if input.Siret != nil {
		store.Siret = *input.Siret
	}

	return s.repo.Update(ctx, store)
}
