package categorie

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
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

func (s *Service) Create(ctx context.Context, catalogID uuid.UUID, input Categorie) (*Categorie, error) {
	if input.Type == "" {
		return nil, errors.New("type is required")
	}
	c := &Categorie{
		Type:      input.Type,
		CatalogID: catalogID,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, catalogID uuid.UUID, input Categorie) (*Categorie, error) {
	c, err := s.repo.Update(ctx, id, catalogID, &input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategorieNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteByID(ctx context.Context, id uuid.UUID, catalogID uuid.UUID) error {
	if err := s.repo.DeleteById(ctx, id, catalogID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategorieNotFound
		}
		return err
	}
	return nil
}

func (s *Service) DeleteByType(ctx context.Context, typ string, catalogID uuid.UUID) error {
	if err := s.repo.DeleteByType(ctx, typ, catalogID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategorieNotFound
		}
		return err
	}
	return nil
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID, catalogID uuid.UUID) (*Categorie, error) {
	c, err := s.repo.FindByID(ctx, id, catalogID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategorieNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) FindByType(ctx context.Context, typ string, catalogID uuid.UUID) (*Categorie, error) {
	c, err := s.repo.FindByType(ctx, typ, catalogID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategorieNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) FindAll(ctx context.Context, catalogID uuid.UUID) ([]Categorie, error) {
	return s.repo.FindAll(ctx, catalogID)
}
