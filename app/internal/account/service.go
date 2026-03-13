package account

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/customer"

	"tili/app/internal/profile"
	"tili/app/internal/store"
)

type LicenseDeleter interface {
	DeleteByAccountID(ctx context.Context, accountID int) error
}

type Service struct {
	repo           *Repository
	storeService   *store.Service
	profileService *profile.Service
	licenseService LicenseDeleter
}

func NewService(repo *Repository, storeService *store.Service, profileService *profile.Service, licenseService LicenseDeleter) *Service {
	return &Service{
		repo:           repo,
		storeService:   storeService,
		profileService: profileService,
		licenseService: licenseService,
	}
}

func (s *Service) Create(ctx context.Context, input RegistrationInput) (*Account, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.FindByEmail(ctx, input.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	params := &stripe.CustomerParams{
		Name:  stripe.String(input.Name),
		Email: stripe.String(input.Email),
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Stripe: %v", err)
	}
	acc := &Account{
		Email:            input.Email,
		Password:         string(hashed),
		Name:             input.Name,
		StripeCustomerID: c.ID,
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*Account, []store.Store, error) {
	acc, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(input.Password)); err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	stores, err := s.storeService.FindByBuyerID(ctx, acc.AccountID)
	if err != nil {
		return nil, nil, err
	}

	return acc, stores, nil
}

func (s *Service) Exists(ctx context.Context, accountID int) (bool, error) {
	_, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *Service) FullDelete(ctx context.Context, id int) error {
	stores, err := s.storeService.FindByBuyerID(ctx, id)
	if err != nil {
		return errors.New("failed to find stores")
	}

	for _, st := range stores {
		if err := s.profileService.DeleteByStoreID(ctx, st.StoreID); err != nil {
			return err
		}
		if err := s.storeService.Delete(ctx, st.StoreID); err != nil {
			return err
		}
	}

	if err := s.licenseService.DeleteByAccountID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*Account, error) {
	return s.repo.FindByID(ctx, id)
}
