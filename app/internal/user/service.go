package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{
		StoreID:      input.StoreID,
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hashed),
		AccessCode:   input.AccessCode,
		AccessLevel:  input.AccessLevel,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) GetAll(ctx context.Context) ([]User, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateUserInput) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		u.Name = input.Name
	}
	if input.Email != "" {
		u.Email = input.Email
	}
	if input.AccessCode != "" {
		u.AccessCode = input.AccessCode
	}
	u.AccessLevel = input.AccessLevel
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) FindByStoreID(ctx context.Context, storeID int64) (*User, error) {
	u, err := s.repo.FindByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}
	return s.repo.Delete(ctx, id)
}
