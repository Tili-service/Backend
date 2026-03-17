package item

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrItemNotFound = errors.New("item not found")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, inputItem Item) (*Item, error) {
	if inputItem.Name == "" {
		return nil, errors.New("name is required")
	}
	if inputItem.Price.IsNegative() {
		return nil, errors.New("price must be a positive number")
	}
	if inputItem.Tax.IsNegative() || inputItem.Tax.GreaterThan(decimal.NewFromFloat(1)) {
		return nil, errors.New("tax must be a positive number between 0 and 1")
	}
	if inputItem.CategorieID <= 0 {
		return nil, errors.New("categorie_id must be a positive integer")
	}
	i := &Item{
		Name:        inputItem.Name,
		Price:       inputItem.Price,
		Tax:         inputItem.Tax,
		CategorieID: inputItem.CategorieID,
	}
	if err := s.repo.Create(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) Update(ctx context.Context, id int, input ItemUpdate) (*Item, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	c, err = s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetByCategorieID(ctx context.Context, id int) ([]Item, error) {
	return s.repo.FindByCategorieID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrItemNotFound
		}
		return err
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]Item, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*Item, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (*Item, error) {
	c, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return c, nil
}
