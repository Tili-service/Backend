package profile

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func generatePIN(length int) (string, error) {
	max := big.NewInt(1)
	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n), nil
}

func (s *Service) generateUniquePin(ctx context.Context, storeID int) (string, error) {
	for i := 0; i < 100; i++ {
		pin, err := generatePIN(6)
		if err != nil {
			return "", err
		}
		exists, err := s.repo.PinExistsInStore(ctx, storeID, pin)
		if err != nil {
			return "", err
		}
		if !exists {
			return pin, nil
		}
	}
	return "", errors.New("failed to generate unique PIN after 100 attempts")
}

func (s *Service) Create(ctx context.Context, input CreateProfileInput) (*ProfileWithPin, error) {
	pin, err := s.generateUniquePin(ctx, input.StoreID)
	if err != nil {
		return nil, err
	}

	levelAccess := input.LevelAccess
	if levelAccess == 0 {
		levelAccess = 4 // default UserLevel
	}

	p := &Profile{
		StoreID:     input.StoreID,
		Name:        input.Name,
		Pin:         pin,
		LevelAccess: levelAccess,
		IsActive:    true,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return &ProfileWithPin{
		ProfileID:   p.ProfileID,
		StoreID:     p.StoreID,
		Name:        p.Name,
		Pin:         pin,
		LevelAccess: p.LevelAccess,
		IsActive:    p.IsActive,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*Profile, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("profile not found")
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) DeleteByStoreID(ctx context.Context, storeID int) error {
	return s.repo.DeleteByStoreID(ctx, storeID)
}

func (s *Service) LoginWithPin(ctx context.Context, storeID int, pin string) (*Profile, error) {
	p, err := s.repo.FindByStoreAndPin(ctx, storeID, pin)
	if err != nil {
		return nil, errors.New("invalid pin")
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id int, input updateProfileInput) (*Profile, error) {
	profile, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	if input.Name != nil {
		profile.Name = *input.Name
	}
	if input.Pin != nil {
		profile.Pin = *input.Pin
	}
	if input.LevelAccess != nil {
		profile.LevelAccess = *input.LevelAccess
	}
	if input.IsActive != nil {
		profile.IsActive = *input.IsActive
	}

	err = s.repo.Update(ctx, profile)
	if err != nil {
		return nil, errors.New("failed to update profile")
	}
	return profile, nil
}
