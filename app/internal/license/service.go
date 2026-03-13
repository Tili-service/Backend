package license

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) DeleteByAccountID(ctx context.Context, accountID int) error {
	return s.repo.DeleteLicencesByAccountID(ctx, accountID)
}

func (s *Service) GetByAccountID(ctx context.Context, accountID int) ([]Licence, error) {
	return s.repo.FindLicencesByAccountID(ctx, accountID)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Licence, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, accountID int, id uuid.UUID) error {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("licence not found")
	}
	if lic.AccountID != accountID {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, accountID int, id uuid.UUID, input UpdateLicenceInput) (*Licence, error) {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("licence not found")
	}
	if lic.AccountID != accountID {
		return nil, errors.New("forbidden")
	}
	if input.Transaction != nil {
		lic.Transaction = *input.Transaction
	}
	if input.IsActive != nil {
		lic.IsActive = *input.IsActive
	}
	return s.repo.Update(ctx, lic)
}

func (s *Service) Create(ctx context.Context, accountID int, input CreateLicenceInput) (*Licence, error) {
	lic := &Licence{
		LicenceID:   uuid.New(),
		AccountID:   accountID,
		Expiration:  time.Now().Add(time.Duration(input.DurationDays) * 24 * time.Hour),
		Transaction: input.Transaction,
		IsActive:    true,
	}
	if err := s.repo.CreateLicence(ctx, lic); err != nil {
		return nil, err
	}
	return lic, nil
}
