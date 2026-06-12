package catalog

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrCatalogNotFound = errors.New("catalog not found")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, storeID int, input catalogUpdate) (*catalog, error) {
	if input.Name == nil || *input.Name == "" {
		return nil, errors.New("name is required")
	}
	c := &catalog{
		Name:    *input.Name,
		StoreID: storeID,
	}
	if input.Description != nil {
		c.Description = *input.Description
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id int, storeID int, input catalogUpdate) (*catalog, error) {
	if (input.Name == nil || *input.Name == "") && (input.Description == nil || *input.Description == "") {
		return nil, errors.New("at least one field is required")
	}
	_, err := s.repo.FindByID(ctx, id, storeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCatalogNotFound
		}
		return nil, err
	}
	c, err := s.repo.Update(ctx, id, storeID, input)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id int, storeID int) error {
	_, err := s.repo.FindByID(ctx, id, storeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCatalogNotFound
		}
		return err
	}
	return s.repo.DeleteByID(ctx, id, storeID)
}

func (s *Service) GetAll(ctx context.Context, storeID int) ([]catalog, error) {
	return s.repo.FindAll(ctx, storeID)
}

func (s *Service) GetByID(ctx context.Context, id int, storeID int) (*catalog, error) {
	c, err := s.repo.FindByID(ctx, id, storeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCatalogNotFound
		}
		return nil, err
	}
	return c, nil
}
