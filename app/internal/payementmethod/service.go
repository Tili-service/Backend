package payementmethod

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrPayementMethodNotFound = errors.New("payement method not found")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input PayementMethod) (*PayementMethod, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	pm := &PayementMethod{
		Name: input.Name,
	}
	if err := s.repo.Create(ctx, pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (s *Service) Update(ctx context.Context, id int, input PayementMethod) (*PayementMethod, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPayementMethodNotFound
		}
		return nil, err
	}
	pm := &PayementMethod{
		PayementMethodID: id,
		Name:             input.Name,
	}
	if err = s.repo.Update(ctx, pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	_, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPayementMethodNotFound
		}
		return err
	}
	return s.repo.DeleteByName(ctx, name)
}

func (s *Service) DeleteByID(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPayementMethodNotFound
		}
		return err
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]PayementMethod, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByName(ctx context.Context, name string) (*PayementMethod, error) {
	pm, err := s.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPayementMethodNotFound
		}
		return nil, err
	}
	return pm, nil
}
