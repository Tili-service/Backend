package categorie

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrCategorieNotFound = errors.New("categorie not found")
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
	c, err := s.repo.Update(ctx, id, &input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategorieNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteByID(ctx context.Context, id int) error {
	if err := s.repo.DeleteById(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategorieNotFound
		}
		return err
	}
	return nil
}

func (s *Service) DeleteByType(ctx context.Context, typ string) error {
	if err := s.repo.DeleteByType(ctx, typ); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategorieNotFound
		}
		return err
	}
	return nil
}

func (s *Service) FindByID(ctx context.Context, id int) (*Categorie, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategorieNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) FindByType(ctx context.Context, typ string) (*Categorie, error) {
	c, err := s.repo.FindByType(ctx, typ)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategorieNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) FindAll(ctx context.Context) ([]Categorie, error) {
	return s.repo.FindAll(ctx)
}
