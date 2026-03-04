package store

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, input CreateStoreInput) (*Store, error) {
	store := &Store{
		StoreName: input.StoreName,
		AccountID: input.AccountID,
	}

	return s.repo.Create(ctx, store)
}

func (s *Service) FindByAccountID(ctx context.Context, accountID int64) (*Store, error) {
	store, err := s.repo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateStoreInput) (*Store, error) {
	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	store.StoreName = input.StoreName
	err = s.repo.Update(ctx, store)
	if err != nil {
		return nil, err
	}
	return store, nil
}
