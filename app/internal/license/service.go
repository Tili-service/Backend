package license

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"

	"tili/app/internal/account"
	"tili/app/internal/store"
	"tili/app/internal/user"
)

type Service struct {
	repo         *Repository
	userService  *user.Service
	storeService *store.Service
}

func NewService(repo *Repository, userService *user.Service, storeService *store.Service) *Service {
	return &Service{
		repo:         repo,
		userService:  userService,
		storeService: storeService,
	}
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

func (s *Service) Create(ctx context.Context, input AccountRegistrationInput) (*bodyResponse, error) {
	account := &account.Account{
		LicenceID: uuid.New(),
		ExpiresAt: time.Now().Add(time.Duration(input.LicenceActive) * 24 * time.Hour),
		IsActive:  true,
	}

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, err
	}

	storeData, err := s.storeService.Create(ctx, store.CreateStoreInput{
		StoreName: input.StoreName,
		AccountID: account.AccountID,
	})
	if err != nil {
		return nil, err
	}

	pin, err := generatePIN(6)
	if err != nil {
		panic(err)
	}
	_, err = s.userService.Create(ctx, user.CreateUserInput{
		StoreID:     storeData.StoreID,
		Name:        input.Name,
		Email:       input.Email,
		Password:    input.Password,
		AccessCode:  pin,
		AccessLevel: 1,
	})
	if err != nil {
		return nil, err
	}
	bodyResponse := bodyResponse{
		AccountID:      account.AccountID,
		UserAccessCode: pin,
		ExpiresAt:      account.ExpiresAt,
	}
	return &bodyResponse, nil
}

func (s *Service) Exists(ctx context.Context, accountID int64) (bool, error) {
	_, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		if err.Error() == "account not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Service) FullDeleteAccount(ctx context.Context, id int64) error {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("account not found")
	}
	storeData, err := s.storeService.FindByAccountID(ctx, account.AccountID)
	if err != nil {
		return errors.New("store not found")
	}
	user, err := s.userService.FindByStoreID(ctx, storeData.StoreID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.userService.Delete(ctx, user.UserID); err != nil {
		return err
	}
	if err := s.storeService.Delete(ctx, storeData.StoreID); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
