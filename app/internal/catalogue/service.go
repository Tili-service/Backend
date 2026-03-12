package catalogue

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CatalogueUpdate) (*Catalogue, error) {
	if input.Name == nil || *input.Name == "" {
		return nil, errors.New("name is required")
	}
	c := &Catalogue{
		Name:        *input.Name,
		Description: *input.Description,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id int, input CatalogueUpdate) (*Catalogue, error) {
	if (input.Name == nil || *input.Name == "") && (input.Description == nil || *input.Description == "") {
		return nil, errors.New("at least one field is required")
	}
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("catalogue not found")
	}
	c, err = s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("catalogue not found")
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]Catalogue, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*Catalogue, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("catalogue not found")
	}
	return c, nil
}
