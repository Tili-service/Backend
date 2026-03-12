package categorie

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

func (s *Service) Create(ctx context.Context, input Categorie) (*Categorie, error) {
	if input.Type == "" {
		return nil, errors.New("type is required")
	}
	c := &Categorie{
		Type: input.Type,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id int, input Categorie) (*Categorie, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("categorie not found")
	}
	c, err = s.repo.Update(ctx, id, &input)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteByID(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("categorie not found")
	}
	return s.repo.DeleteById(ctx, id)
}

func (s *Service) DeleteByType(ctx context.Context, typ string) error {
	_, err := s.repo.FindByType(ctx, typ)
	if err != nil {
		return errors.New("categorie not found")
	}
	return s.repo.DeleteByType(ctx, typ)
}

func (s *Service) FindAll(ctx context.Context) ([]Categorie, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) FindByID(ctx context.Context, id int) (*Categorie, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) FindByType(ctx context.Context, typ string) (*Categorie, error) {
	return s.repo.FindByType(ctx, typ)
}
